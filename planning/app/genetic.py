from __future__ import annotations

import random
import time
from collections.abc import Sequence

from deap import algorithms, base, creator, tools

from .fitness import score_plan
from .models import CandidateLesson, Conflict, LessonOption, PlanningContext


DEFAULT_SEED = 42
DEFAULT_POPULATION_SIZE = 100
DEFAULT_GENERATIONS = 50
DEFAULT_CROSSOVER_PROBABILITY = 0.8
DEFAULT_MUTATION_PROBABILITY = 0.3
DEFAULT_GENE_MUTATION_PROBABILITY = 0.3
DEFAULT_TOURNAMENT_SIZE = 3
DEFAULT_MAX_SECONDS = 60.0


def evolve_plan(
    context: PlanningContext,
    seed: int = DEFAULT_SEED,
    population_size: int = DEFAULT_POPULATION_SIZE,
    generations: int = DEFAULT_GENERATIONS,
    max_seconds: float = DEFAULT_MAX_SECONDS,
    verbose: bool = False,
) -> tuple[list[CandidateLesson], list[str], list[Conflict], float]:
    if not context.lesson_demands:
        return [], [], [], score_plan([], [], [], context)

    if not context.valid_options_by_student:
        unassigned = _unique(context.lesson_demands)
        conflicts = [Conflict(student_id=student_id, reason="no_available_option") for student_id in unassigned]
        return [], unassigned, conflicts, score_plan([], unassigned, conflicts, context)

    ga_started_at = time.monotonic()
    _ensure_creators()
    random.seed(seed)
    rng = random.Random(seed)
    toolbox = base.Toolbox()
    student_index_by_id = {student_id: index for index, student_id in enumerate(context.students_to_plan)}
    greedy_individual = _greedy_individual(context, student_index_by_id)
    greedy_lessons, greedy_unassigned_students, greedy_conflicts = decode_individual(greedy_individual, context)
    greedy_score = score_plan(greedy_lessons, greedy_unassigned_students, greedy_conflicts, context)
    optimize_complete_plan = (
        len(greedy_lessons) == len(context.lesson_demands)
        and not greedy_unassigned_students
        and not greedy_conflicts
    )

    def init_individual() -> creator.PlanningIndividual:
        genes: list[int] = []
        for student_id in context.lesson_demands:
            student_index = student_index_by_id.get(student_id, 0)
            option_count = len(context.valid_options_by_student.get(student_id, []))
            option_index = rng.randint(0, option_count - 1) if option_count else -1
            genes.extend([student_index, option_index])
        return creator.PlanningIndividual(genes)

    toolbox.register("individual", init_individual)
    toolbox.register("population", tools.initRepeat, list, toolbox.individual)
    toolbox.register("mate", tools.cxTwoPoint)
    if optimize_complete_plan:
        toolbox.register("mutate", _mutate_complete_plan_individual, rng=rng, context=context)
    else:
        toolbox.register(
            "mutate",
            _mutate_individual,
            rng=rng,
            context=context,
            indpb=DEFAULT_GENE_MUTATION_PROBABILITY,
        )
    toolbox.register("select", tools.selTournament, tournsize=DEFAULT_TOURNAMENT_SIZE)
    toolbox.register(
        "evaluate",
        _evaluate_individual,
        context=context,
        complete_plan_floor=greedy_score if optimize_complete_plan else None,
    )

    if optimize_complete_plan:
        population = _complete_plan_population(greedy_individual, population_size, toolbox)
    else:
        population = toolbox.population(n=population_size)
        if population:
            population[0] = greedy_individual
    if population:
        population[0].fitness.values = toolbox.evaluate(population[0])
        if verbose:
            print(
                "[planning] greedy baseline: "
                f"lessons={len(greedy_lessons)} "
                f"unassigned_students={len(greedy_unassigned_students)} "
                f"conflicts={len(greedy_conflicts)} "
                f"score={greedy_score:.2f} "
                f"mode={'complete_plan_optimizer' if optimize_complete_plan else 'repairing_search'}",
                flush=True,
            )
    hall_of_fame = tools.HallOfFame(1)
    stats = tools.Statistics(lambda ind: ind.fitness.values[0])
    stats.register("max", max)
    stats.register("avg", lambda scores: sum(scores) / len(scores))

    remaining_seconds = max_seconds - (time.monotonic() - ga_started_at)
    _run_elitist_evolution(
        population,
        toolbox,
        cxpb=0.0 if optimize_complete_plan else DEFAULT_CROSSOVER_PROBABILITY,
        mutpb=DEFAULT_MUTATION_PROBABILITY,
        generations=generations,
        max_seconds=remaining_seconds,
        stats=stats,
        hall_of_fame=hall_of_fame,
        verbose=verbose,
    )

    best_individual = hall_of_fame[0] if hall_of_fame else max(population, key=lambda ind: ind.fitness.values[0])
    lessons, unassigned_students, conflicts = decode_individual(best_individual, context)
    final_score = score_plan(lessons, unassigned_students, conflicts, context)
    return lessons, unassigned_students, conflicts, final_score


def decode_individual(
    individual: Sequence[int],
    context: PlanningContext,
) -> tuple[list[CandidateLesson], list[str], list[Conflict]]:
    lessons: list[CandidateLesson] = []
    unassigned_students: list[str] = []
    conflicts: list[Conflict] = []
    occupied_students, occupied_instructors, occupied_cars = _occupied_sets(context)

    for demand_index in range(len(context.lesson_demands)):
        gene_offset = demand_index * 2
        option_index = individual[gene_offset + 1]
        student_id = context.lesson_demands[demand_index]
        option, selected_candidate = _select_non_conflicting_option(
            student_id,
            option_index,
            occupied_students,
            occupied_instructors,
            occupied_cars,
            context,
        )
        if option is None or selected_candidate is None:
            _append_unique(unassigned_students, student_id)
            conflicts.append(Conflict(student_id=student_id, reason="no_available_option"))
            continue

        lessons.append(selected_candidate)
        _occupy(selected_candidate, occupied_students, occupied_instructors, occupied_cars)

    lessons.sort(key=lambda lesson: (lesson.start_at, lesson.student_id, lesson.instructor_id, lesson.car_id))
    return lessons, unassigned_students, conflicts


def _select_non_conflicting_option(
    student_id: str,
    option_index: int,
    occupied_students: set[tuple[object, str]],
    occupied_instructors: set[tuple[object, str]],
    occupied_cars: set[tuple[object, str]],
    context: PlanningContext,
) -> tuple[LessonOption | None, CandidateLesson | None]:
    options = context.valid_options_by_student.get(student_id, [])
    if not options or option_index < 0:
        return None, None

    start_index = option_index % len(options)
    for offset in range(len(options)):
        option = options[(start_index + offset) % len(options)]
        candidate = _candidate_from_option(option)
        if _has_indexed_conflict(candidate, occupied_students, occupied_instructors, occupied_cars):
            continue
        return option, candidate
    return None, None


def _option_for_gene(student_id: str, option_index: int, context: PlanningContext) -> LessonOption | None:
    options = context.valid_options_by_student.get(student_id, [])
    if not options or option_index < 0:
        return None
    return options[option_index % len(options)]


def _candidate_from_option(option: LessonOption) -> CandidateLesson:
    return CandidateLesson(
        student_id=option.student_id,
        instructor_id=option.instructor_id,
        car_id=option.car_id,
        start_at=option.time_slot.start_at,
        end_at=option.time_slot.end_at,
    )


def _append_unique(items: list[str], value: str) -> None:
    if value not in items:
        items.append(value)


def _unique(items: Sequence[str]) -> list[str]:
    result: list[str] = []
    for item in items:
        _append_unique(result, item)
    return result


def _evaluate_individual(
    individual: Sequence[int],
    context: PlanningContext,
    complete_plan_floor: float | None = None,
) -> tuple[float]:
    lessons, unassigned_students, conflicts = decode_individual(individual, context)
    if complete_plan_floor is not None and (
        len(lessons) != len(context.lesson_demands)
        or unassigned_students
        or conflicts
    ):
        missing_count = max(0, len(context.lesson_demands) - len(lessons))
        return (complete_plan_floor - 1_000_000.0 - missing_count * 10_000.0 - len(conflicts) * 10_000.0,)
    return (score_plan(lessons, unassigned_students, conflicts, context),)


def _run_elitist_evolution(
    population: list[creator.PlanningIndividual],
    toolbox: base.Toolbox,
    cxpb: float,
    mutpb: float,
    generations: int,
    max_seconds: float,
    stats: tools.Statistics,
    hall_of_fame: tools.HallOfFame,
    verbose: bool,
) -> None:
    started_at = time.monotonic()
    deadline = started_at + max(0.0, max_seconds)
    logbook = tools.Logbook()
    logbook.header = ["gen", "nevals"] + (stats.fields if stats else [])

    invalid_individuals = [individual for individual in population if not individual.fitness.valid]
    nevals, timed_out = _evaluate_invalid_individuals(invalid_individuals, toolbox, deadline)

    valid_population = [individual for individual in population if individual.fitness.valid]
    if not valid_population:
        if verbose:
            print("[planning] genetic algorithm stopped: no evaluated individuals before timeout", flush=True)
        return

    hall_of_fame.update(valid_population)
    record = stats.compile(valid_population) if stats else {}
    logbook.record(gen=0, nevals=nevals, **record)
    if verbose:
        print(logbook.stream, flush=True)

    if not population or timed_out:
        if timed_out and verbose:
            print(
                f"[planning] genetic algorithm stopped by time limit: elapsed={time.monotonic() - started_at:.1f}s",
                flush=True,
            )
        return

    elite_count = 1
    for generation in range(1, generations + 1):
        if time.monotonic() >= deadline:
            if verbose:
                print(
                    f"[planning] genetic algorithm stopped by time limit: elapsed={time.monotonic() - started_at:.1f}s",
                    flush=True,
                )
            return

        elites = [toolbox.clone(individual) for individual in tools.selBest(population, elite_count)]
        offspring = toolbox.select(population, max(0, len(population) - elite_count))
        offspring = algorithms.varAnd(list(map(toolbox.clone, offspring)), toolbox, cxpb, mutpb)

        invalid_individuals = [individual for individual in offspring if not individual.fitness.valid]
        nevals, timed_out = _evaluate_invalid_individuals(invalid_individuals, toolbox, deadline)
        if timed_out:
            if verbose:
                print(
                    f"[planning] genetic algorithm stopped by time limit: elapsed={time.monotonic() - started_at:.1f}s",
                    flush=True,
                )
            return

        population[:] = elites + offspring
        hall_of_fame.update(population)
        record = stats.compile(population) if stats else {}
        logbook.record(gen=generation, nevals=nevals, **record)
        if verbose:
            print(logbook.stream, flush=True)


def _evaluate_invalid_individuals(
    individuals: list[creator.PlanningIndividual],
    toolbox: base.Toolbox,
    deadline: float,
) -> tuple[int, bool]:
    nevals = 0
    for individual in individuals:
        if time.monotonic() >= deadline:
            return nevals, True
        individual.fitness.values = toolbox.evaluate(individual)
        nevals += 1
    return nevals, False


def _complete_plan_population(
    greedy_individual: creator.PlanningIndividual,
    population_size: int,
    toolbox: base.Toolbox,
) -> list[creator.PlanningIndividual]:
    if population_size <= 0:
        return []
    population = [creator.PlanningIndividual(greedy_individual)]
    for _index in range(1, population_size):
        individual = creator.PlanningIndividual(greedy_individual)
        toolbox.mutate(individual)
        population.append(individual)
    return population


def _greedy_individual(
    context: PlanningContext,
    student_index_by_id: dict[str, int],
) -> creator.PlanningIndividual:
    genes: list[int] = []
    occupied_students, occupied_instructors, occupied_cars = _occupied_sets(context)
    for student_id in context.lesson_demands:
        options = context.valid_options_by_student.get(student_id, [])
        selected_index = -1
        for option_index, option in enumerate(options):
            candidate = _candidate_from_option(option)
            if _has_indexed_conflict(candidate, occupied_students, occupied_instructors, occupied_cars):
                continue
            selected_index = option_index
            _occupy(candidate, occupied_students, occupied_instructors, occupied_cars)
            break
        genes.extend([student_index_by_id.get(student_id, 0), selected_index])
    return creator.PlanningIndividual(genes)


def _mutate_individual(
    individual: list[int],
    rng: random.Random,
    context: PlanningContext,
    indpb: float,
) -> tuple[list[int]]:
    for offset in range(0, len(individual), 2):
        if rng.random() >= indpb:
            continue
        demand_index = offset // 2
        student_id = context.lesson_demands[demand_index]
        option_count = len(context.valid_options_by_student.get(student_id, []))
        individual[offset + 1] = rng.randint(0, option_count - 1) if option_count else -1
    return (individual,)


def _mutate_complete_plan_individual(
    individual: list[int],
    rng: random.Random,
    context: PlanningContext,
) -> tuple[list[int]]:
    if not individual:
        return (individual,)

    mutation_count = rng.randint(1, 3)
    for _index in range(mutation_count):
        _try_preserving_mutation(individual, rng, context)
    return (individual,)


def _try_preserving_mutation(
    individual: list[int],
    rng: random.Random,
    context: PlanningContext,
) -> None:
    demand_count = len(context.lesson_demands)
    if demand_count <= 0:
        return

    demand_index = rng.randrange(demand_count)
    offset = demand_index * 2
    student_id = context.lesson_demands[demand_index]
    option_count = len(context.valid_options_by_student.get(student_id, []))
    if option_count <= 1:
        return

    original_option_index = individual[offset + 1]
    for _attempt in range(16):
        candidate_option_index = rng.randrange(option_count)
        if candidate_option_index == original_option_index:
            continue
        individual[offset + 1] = candidate_option_index
        lessons, unassigned_students, conflicts = decode_individual(individual, context)
        if len(lessons) == demand_count and not unassigned_students and not conflicts:
            return

    individual[offset + 1] = original_option_index


def _ensure_creators() -> None:
    if not hasattr(creator, "PlanningFitnessMax"):
        creator.create("PlanningFitnessMax", base.Fitness, weights=(1.0,))
    if not hasattr(creator, "PlanningIndividual"):
        creator.create("PlanningIndividual", list, fitness=creator.PlanningFitnessMax)


def _occupied_sets(context: PlanningContext) -> tuple[set[tuple[object, str]], set[tuple[object, str]], set[tuple[object, str]]]:
    occupied_students = {
        (lesson.start_at, lesson.student_id)
        for lesson in context.existing_lessons
    }
    occupied_instructors = {
        (lesson.start_at, lesson.instructor_id)
        for lesson in context.existing_lessons
    }
    occupied_cars = {
        (lesson.start_at, lesson.car_id)
        for lesson in context.existing_lessons
    }
    return occupied_students, occupied_instructors, occupied_cars


def _has_indexed_conflict(
    candidate: CandidateLesson,
    occupied_students: set[tuple[object, str]],
    occupied_instructors: set[tuple[object, str]],
    occupied_cars: set[tuple[object, str]],
) -> bool:
    return (
        (candidate.start_at, candidate.student_id) in occupied_students
        or (candidate.start_at, candidate.instructor_id) in occupied_instructors
        or (candidate.start_at, candidate.car_id) in occupied_cars
    )


def _occupy(
    candidate: CandidateLesson,
    occupied_students: set[tuple[object, str]],
    occupied_instructors: set[tuple[object, str]],
    occupied_cars: set[tuple[object, str]],
) -> None:
    occupied_students.add((candidate.start_at, candidate.student_id))
    occupied_instructors.add((candidate.start_at, candidate.instructor_id))
    occupied_cars.add((candidate.start_at, candidate.car_id))

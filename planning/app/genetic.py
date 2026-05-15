from __future__ import annotations

import random
from collections.abc import Sequence

from deap import algorithms, base, creator, tools

from .constraints import has_conflict, transmission_matches
from .fitness import score_plan
from .models import CandidateLesson, Conflict, PlanningContext


DEFAULT_SEED = 42
DEFAULT_POPULATION_SIZE = 100
DEFAULT_GENERATIONS = 50
DEFAULT_CROSSOVER_PROBABILITY = 0.8
DEFAULT_MUTATION_PROBABILITY = 0.3
DEFAULT_GENE_MUTATION_PROBABILITY = 0.3
DEFAULT_TOURNAMENT_SIZE = 3


def evolve_plan(
    slots: Sequence[tuple],
    context: PlanningContext,
    seed: int = DEFAULT_SEED,
    population_size: int = DEFAULT_POPULATION_SIZE,
    generations: int = DEFAULT_GENERATIONS,
) -> tuple[list[CandidateLesson], list[str], list[Conflict], float]:
    if not context.students_to_plan or not context.instructor_ids or not context.car_ids or not slots:
        unassigned = list(context.students_to_plan)
        conflicts = [Conflict(student_id=student_id, reason="no_available_slot") for student_id in unassigned]
        return [], unassigned, conflicts, score_plan([], unassigned, conflicts, context)

    _ensure_creators()
    random.seed(seed)
    rng = random.Random(seed)
    toolbox = base.Toolbox()

    slot_max = len(slots) - 1
    instructor_max = len(context.instructor_ids) - 1
    car_max = len(context.car_ids) - 1

    def init_individual() -> creator.PlanningIndividual:
        genes: list[int] = []
        for _student_id in context.students_to_plan:
            genes.extend(
                [
                    rng.randint(-1, slot_max),
                    rng.randint(0, instructor_max),
                    rng.randint(0, car_max),
                ]
            )
        return creator.PlanningIndividual(genes)

    toolbox.register("individual", init_individual)
    toolbox.register("population", tools.initRepeat, list, toolbox.individual)
    toolbox.register("mate", tools.cxTwoPoint)
    toolbox.register(
        "mutate",
        _mutate_individual,
        rng=rng,
        slot_max=slot_max,
        instructor_max=instructor_max,
        car_max=car_max,
        indpb=DEFAULT_GENE_MUTATION_PROBABILITY,
    )
    toolbox.register("select", tools.selTournament, tournsize=DEFAULT_TOURNAMENT_SIZE)
    toolbox.register("evaluate", _evaluate_individual, context=context, slots=slots)

    population = toolbox.population(n=population_size)
    hall_of_fame = tools.HallOfFame(1)
    stats = tools.Statistics(lambda ind: ind.fitness.values[0])
    stats.register("max", max)
    stats.register("avg", lambda scores: sum(scores) / len(scores))

    algorithms.eaSimple(
        population,
        toolbox,
        cxpb=DEFAULT_CROSSOVER_PROBABILITY,
        mutpb=DEFAULT_MUTATION_PROBABILITY,
        ngen=generations,
        stats=stats,
        halloffame=hall_of_fame,
        verbose=False,
    )

    best_individual = hall_of_fame[0] if hall_of_fame else max(population, key=lambda ind: ind.fitness.values[0])
    lessons, unassigned_students, conflicts = decode_individual(best_individual, context, slots)
    final_score = score_plan(lessons, unassigned_students, conflicts, context)
    return lessons, unassigned_students, conflicts, final_score


def decode_individual(
    individual: Sequence[int],
    context: PlanningContext,
    slots: Sequence[tuple],
) -> tuple[list[CandidateLesson], list[str], list[Conflict]]:
    lessons: list[CandidateLesson] = []
    unassigned_students: list[str] = []
    conflicts: list[Conflict] = []

    for index, student_id in enumerate(context.students_to_plan):
        gene_offset = index * 3
        slot_index = individual[gene_offset]
        instructor_index = individual[gene_offset + 1]
        car_index = individual[gene_offset + 2]

        if slot_index < 0 or slot_index >= len(slots):
            unassigned_students.append(student_id)
            conflicts.append(Conflict(student_id=student_id, reason="no_available_slot"))
            continue

        start_at, end_at = slots[slot_index]
        candidate = CandidateLesson(
            student_id=student_id,
            instructor_id=context.instructor_ids[instructor_index % len(context.instructor_ids)],
            car_id=context.car_ids[car_index % len(context.car_ids)],
            start_at=start_at,
            end_at=end_at,
        )
        if not transmission_matches(
            context.student_transmissions.get(student_id, ""),
            context.car_transmissions.get(candidate.car_id, ""),
        ):
            unassigned_students.append(student_id)
            conflicts.append(Conflict(student_id=student_id, reason="transmission_mismatch"))
            continue
        if has_conflict(candidate, lessons, context.existing_lessons):
            unassigned_students.append(student_id)
            conflicts.append(Conflict(student_id=student_id, reason="resource_conflict"))
            continue

        lessons.append(candidate)

    return lessons, unassigned_students, conflicts


def _evaluate_individual(
    individual: Sequence[int],
    context: PlanningContext,
    slots: Sequence[tuple],
) -> tuple[float]:
    lessons, unassigned_students, conflicts = decode_individual(individual, context, slots)
    return (score_plan(lessons, unassigned_students, conflicts, context),)


def _mutate_individual(
    individual: list[int],
    rng: random.Random,
    slot_max: int,
    instructor_max: int,
    car_max: int,
    indpb: float,
) -> tuple[list[int]]:
    for offset in range(0, len(individual), 3):
        if rng.random() < indpb:
            individual[offset] = rng.randint(-1, slot_max)
        if rng.random() < indpb:
            individual[offset + 1] = rng.randint(0, instructor_max)
        if rng.random() < indpb:
            individual[offset + 2] = rng.randint(0, car_max)
    return (individual,)


def _ensure_creators() -> None:
    if not hasattr(creator, "PlanningFitnessMax"):
        creator.create("PlanningFitnessMax", base.Fitness, weights=(1.0,))
    if not hasattr(creator, "PlanningIndividual"):
        creator.create("PlanningIndividual", list, fitness=creator.PlanningFitnessMax)

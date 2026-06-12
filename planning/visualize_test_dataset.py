from __future__ import annotations

import argparse
import json
from pathlib import Path
from typing import Any


DEFAULT_DATASET = Path(__file__).resolve().parent / "generated" / "planning_test_dataset.json"
DEFAULT_REPORT = Path(__file__).resolve().parent / "generated" / "planning_run_report.json"
DEFAULT_OUTPUT = Path(__file__).resolve().parent / "generated" / "planning_visualization.html"


def main() -> None:
    args = _parse_args()
    dataset = json.loads(args.dataset.read_text(encoding="utf-8"))
    if args.report.exists():
        report = json.loads(args.report.read_text(encoding="utf-8"))
    else:
        print(f"[visualization] report not found, using dataset reference schedule: {args.report}", flush=True)
        report = _reference_report(dataset)
    html_doc = _render_html(dataset, report)
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(html_doc, encoding="utf-8")
    print(f"[visualization] html saved: {args.output}", flush=True)


def _parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Build an HTML visualization for the planning test dataset.")
    parser.add_argument("--dataset", type=Path, default=DEFAULT_DATASET)
    parser.add_argument("--report", type=Path, default=DEFAULT_REPORT)
    parser.add_argument("--output", type=Path, default=DEFAULT_OUTPUT)
    return parser.parse_args()


def _render_html(dataset: dict[str, Any], report: dict[str, Any]) -> str:
    payload = _json_script({"dataset": dataset, "report": report})
    return f"""<!doctype html>
<html lang="ru">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Planning Dataset Visualization</title>
  <style>
    :root {{
      color-scheme: light;
      --bg: #f6f7f9;
      --panel: #ffffff;
      --text: #20242a;
      --muted: #68717d;
      --line: #d9dee6;
      --manual: #2d6cdf;
      --automatic: #0b8f6a;
      --existing: #d88922;
      --idle: #d7dde6;
      --danger: #c73838;
    }}
    * {{ box-sizing: border-box; }}
    body {{
      margin: 0;
      background: var(--bg);
      color: var(--text);
      font-family: Inter, Segoe UI, Arial, sans-serif;
      line-height: 1.35;
    }}
    header {{
      padding: 24px 32px 16px;
      background: var(--panel);
      border-bottom: 1px solid var(--line);
    }}
    h1 {{ margin: 0 0 8px; font-size: 26px; letter-spacing: 0; }}
    h2 {{ margin: 0 0 14px; font-size: 18px; letter-spacing: 0; }}
    .sub {{ color: var(--muted); font-size: 14px; }}
    main {{ padding: 20px 32px 40px; display: grid; gap: 18px; }}
    section {{
      background: var(--panel);
      border: 1px solid var(--line);
      border-radius: 8px;
      padding: 18px;
    }}
    .metrics {{
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
      gap: 10px;
    }}
    .metric {{
      border: 1px solid var(--line);
      border-radius: 8px;
      padding: 12px;
      min-height: 78px;
    }}
    .metric .label {{ color: var(--muted); font-size: 12px; }}
    .metric .value {{ margin-top: 6px; font-size: 24px; font-weight: 700; }}
    .grid2 {{ display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1fr); gap: 18px; }}
    .bars {{ display: grid; gap: 9px; }}
    .bar-row {{ display: grid; grid-template-columns: 150px 1fr 80px; gap: 10px; align-items: center; font-size: 13px; }}
    .bar-label {{ overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }}
    .bar-track {{ height: 18px; background: var(--idle); border-radius: 4px; overflow: hidden; display: flex; }}
    .bar-gen {{ background: var(--manual); }}
    .bar-existing {{ background: var(--existing); }}
    .bar-auto {{ background: var(--automatic); }}
    .legend {{ display: flex; gap: 14px; flex-wrap: wrap; color: var(--muted); font-size: 12px; margin-bottom: 12px; }}
    .swatch {{ display: inline-block; width: 10px; height: 10px; border-radius: 2px; margin-right: 5px; vertical-align: -1px; }}
    .calendar {{ display: grid; grid-template-columns: repeat(auto-fill, minmax(44px, 1fr)); gap: 6px; }}
    .day {{
      min-height: 46px;
      border: 1px solid var(--line);
      border-radius: 6px;
      padding: 5px;
      font-size: 11px;
      display: grid;
      align-content: space-between;
    }}
    .day strong {{ font-size: 13px; }}
    table {{ width: 100%; border-collapse: collapse; font-size: 13px; }}
    th, td {{ padding: 8px 10px; border-bottom: 1px solid var(--line); text-align: left; }}
    th {{ color: var(--muted); font-weight: 600; background: #fafbfc; position: sticky; top: 0; }}
    .table-wrap {{ max-height: 520px; overflow: auto; border: 1px solid var(--line); border-radius: 8px; }}
    .pill {{ display: inline-block; padding: 2px 7px; border-radius: 999px; font-size: 12px; background: #e8edf5; }}
    .pill.auto {{ background: #dff4ee; color: #09664d; }}
    .pill.manual {{ background: #e6eefc; color: #204f9e; }}
    .score-grid {{ display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 8px 18px; }}
    .score-row {{ display: grid; grid-template-columns: minmax(0, 1fr) auto; align-items: baseline; border-bottom: 1px solid var(--line); padding: 6px 0; gap: 12px; }}
    .score-row span {{ min-width: 0; overflow-wrap: anywhere; }}
    @media (max-width: 900px) {{
      header, main {{ padding-left: 14px; padding-right: 14px; }}
      .grid2 {{ grid-template-columns: 1fr; }}
      .bar-row {{ grid-template-columns: 120px 1fr 64px; }}
    }}
  </style>
</head>
<body>
  <header>
    <h1>Planning Dataset Visualization</h1>
    <div class="sub" id="period"></div>
  </header>
  <main>
    <section>
      <h2>Сводка</h2>
      <div class="metrics" id="metrics"></div>
    </section>
    <section class="grid2">
      <div>
        <h2>Загрузка автомобилей</h2>
        <div class="legend"><span><i class="swatch" style="background:var(--manual)"></i>planning</span><span><i class="swatch" style="background:var(--existing)"></i>существующие</span></div>
        <div class="bars" id="carBars"></div>
      </div>
      <div>
        <h2>Загрузка инструкторов</h2>
        <div class="legend"><span><i class="swatch" style="background:var(--manual)"></i>planning</span><span><i class="swatch" style="background:var(--existing)"></i>существующие</span></div>
        <div class="bars" id="instructorBars"></div>
      </div>
    </section>
    <section>
      <h2>Календарь сгенерированных занятий</h2>
      <div class="calendar" id="calendar"></div>
    </section>
    <section>
      <h2>Лучший fitness score и разбивка</h2>
      <div class="score-grid" id="score"></div>
    </section>
    <section>
      <h2>Студенты</h2>
      <div class="table-wrap">
        <table>
          <thead><tr><th>Студент</th><th>Тип</th><th>Цель</th><th>Поставлено</th><th>Осталось</th><th>Последнее занятие</th><th>Предпочтения</th></tr></thead>
          <tbody id="students"></tbody>
        </table>
      </div>
    </section>
    <section>
      <h2>Расписание planning</h2>
      <div class="table-wrap">
        <table>
          <thead><tr><th>Дата</th><th>Время</th><th>Студент</th><th>Инструктор</th><th>Авто</th><th>Статус</th></tr></thead>
          <tbody id="lessons"></tbody>
        </table>
      </div>
    </section>
  </main>
  <script id="payload" type="application/json">{payload}</script>
  <script>
    const data = JSON.parse(document.getElementById('payload').textContent);
    const report = data.report;
    const dataset = data.dataset;
    const summary = report.summary;
    const fmt = new Intl.NumberFormat('ru-RU', {{ maximumFractionDigits: 1 }});

    document.getElementById('period').textContent = `${{report.request.period_start}} - ${{report.request.period_end}}`;
    metricGrid('metrics', [
      ['Студентов', summary.students],
      ['Автомат', summary.automatic_students],
      ['Авто', summary.cars],
      ['Инструкторов', summary.instructors],
      ['Существующих занятий', summary.existing_lessons],
      ['Целевых занятий', summary.target_lessons],
      ['Поставлено planning', summary.generated_lessons],
      ['Незакрытых студентов', summary.unassigned_students],
      ['Лучший fitness score', fmt.format(summary.best_fitness_score ?? summary.score)],
      ['Доступно авто, ч', fmt.format(summary.car_available_hours ?? 0)],
      ['Простой авто, ч', fmt.format(summary.car_idle_hours ?? 0)],
      ['Доступно инструкторов, ч', fmt.format(summary.instructor_available_hours ?? 0)],
      ['Простой инструкторов, ч', fmt.format(summary.instructor_idle_hours ?? 0)]
    ]);

    function metricGrid(id, items) {{
      document.getElementById(id).innerHTML = items.map(([label, value]) => `
        <div class="metric"><div class="label">${{escapeHtml(label)}}</div><div class="value">${{escapeHtml(String(value))}}</div></div>
      `).join('');
    }}

    function bars(id, rows, labelKey, maxValue, colorByTransmission) {{
      document.getElementById(id).innerHTML = rows.map(row => {{
        const generated = row.generated_lessons;
        const existing = row.existing_lessons;
        const total = generated + existing;
        const availableHours = row.available_hours || row.busy_hours || 0;
        const busyHours = row.busy_hours || 0;
        const genWidth = availableHours ? generated * 1.5 / availableHours * 100 : 0;
        const existingWidth = availableHours ? existing * 1.5 / availableHours * 100 : 0;
        const genClass = colorByTransmission && row.transmission === 'automatic' ? 'bar-gen bar-auto' : 'bar-gen';
        return `<div class="bar-row">
          <div class="bar-label" title="${{escapeHtml(row[labelKey])}}">${{escapeHtml(row[labelKey])}}</div>
          <div class="bar-track"><div class="${{genClass}}" style="width:${{genWidth}}%"></div><div class="bar-existing" style="width:${{existingWidth}}%"></div></div>
          <div>${{fmt.format(busyHours)}}/${{fmt.format(availableHours)}}ч</div>
        </div>`;
      }}).join('');
    }}

    function renderCalendar() {{
      const values = report.lessons_by_date;
      const max = Math.max(...Object.values(values), 1);
      document.getElementById('calendar').innerHTML = Object.entries(values).map(([day, count]) => {{
        const alpha = 0.12 + count / max * 0.78;
        return `<div class="day" style="background:rgba(45,108,223,${{alpha}})"><span>${{day.slice(5)}}</span><strong>${{count}}</strong></div>`;
      }}).join('');
    }}

    const scoreLabels = {{
      score: 'Итоговый score',
      scheduled_lessons: 'Поставлено занятий',
      scheduled_bonus: 'Бонус за поставленные занятия',
      missing_target_lessons: 'Недостающие занятия',
      unmet_lesson_penalty: 'Штраф за незакрытый demand',
      conflicts: 'Конфликты',
      conflict_penalty: 'Штраф за конфликты',
      idle_penalty: 'Штраф за неиспользованную емкость',
      unevenness_penalty: 'Штраф за неравномерность студентов',
      resource_unevenness_penalty: 'Штраф за неравномерность ресурсов',
      preference_bonus: 'Бонус за предпочтения',
      scheduled_lesson_bonus_weight: 'Вес бонуса за занятие',
      missing_lesson_penalty_weight: 'Вес штрафа за недостающее занятие',
      conflict_penalty_weight: 'Вес штрафа за конфликт',
      idle_capacity_penalty_weight: 'Вес штрафа за простой емкости',
      student_unevenness_penalty_weight: 'Вес неравномерности студентов',
      resource_unevenness_penalty_weight: 'Вес неравномерности ресурсов',
      preference_match_bonus_weight: 'Вес совпадения предпочтения',
      preference_miss_penalty_weight: 'Вес несовпадения предпочтения',
      dataset_existing_lessons: 'Существующие занятия в датасете',
      dataset_students: 'Студентов в датасете',
      dataset_cars: 'Автомобилей в датасете',
      dataset_instructors: 'Инструкторов в датасете',
    }};

    function renderScore() {{
      const breakdown = report.fitness_score_breakdown || report.score_breakdown;
      if (!breakdown || typeof breakdown !== 'object' || Array.isArray(breakdown)) {{
        document.getElementById('score').innerHTML = `
          <div class="score-row"><span>Разбивка score</span><strong>${{escapeHtml(String(breakdown || 'нет данных'))}}</strong></div>
        `;
        return;
      }}
      document.getElementById('score').innerHTML = Object.entries(breakdown).map(([key, value]) => `
        <div class="score-row"><span>${{escapeHtml(scoreLabels[key] || key)}}</span><strong>${{typeof value === 'number' ? fmt.format(value) : escapeHtml(String(value))}}</strong></div>
      `).join('');
    }}

    function renderStudents() {{
      document.getElementById('students').innerHTML = report.students.map(row => `
        <tr>
          <td>${{escapeHtml(row.id)}}</td>
          <td><span class="pill ${{row.transmission}}">${{escapeHtml(row.transmission)}}</span></td>
          <td>${{row.target_lessons}}</td>
          <td>${{row.generated_lessons}}</td>
          <td>${{row.remaining_lessons}}</td>
          <td>${{escapeHtml(row.last_lesson_at || '-')}}</td>
          <td>${{escapeHtml(row.preferred_weekdays.join(','))}} / ${{escapeHtml(row.preferred_time_windows.join(','))}}</td>
        </tr>
      `).join('');
    }}

    function renderLessons() {{
      const carById = new Map(dataset.cars.map(car => [car.id, car]));
      document.getElementById('lessons').innerHTML = report.generated_lessons.map(row => {{
        const car = carById.get(row.car_id);
        const start = new Date(row.start_at);
        const end = new Date(row.end_at);
        return `<tr>
          <td>${{row.start_at.slice(0, 10)}}</td>
          <td>${{row.start_at.slice(11, 16)}}-${{row.end_at.slice(11, 16)}}</td>
          <td>${{escapeHtml(row.student_id)}}</td>
          <td>${{escapeHtml(row.instructor_id)}}</td>
          <td>${{escapeHtml(row.car_id)}} <span class="pill ${{car?.transmission || ''}}">${{escapeHtml(car?.transmission || '')}}</span></td>
          <td>${{escapeHtml(row.status)}}</td>
        </tr>`;
      }}).join('');
    }}

    function escapeHtml(value) {{
      return value.replace(/[&<>"']/g, ch => ({{ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }}[ch]));
    }}

    bars('carBars', report.cars, 'id', Math.max(...report.cars.map(x => x.generated_lessons + x.existing_lessons)), true);
    bars('instructorBars', report.instructors, 'id', Math.max(...report.instructors.map(x => x.generated_lessons + x.existing_lessons)), false);
    renderCalendar();
    renderScore();
    renderStudents();
    renderLessons();
  </script>
</body>
</html>"""


def _reference_report(dataset: dict[str, Any]) -> dict[str, Any]:
    generated_lessons = list(dataset["generated_lessons"])
    existing_lessons = list(dataset["existing_lessons"])
    students = list(dataset["students"])
    cars = list(dataset["cars"])
    instructors = list(dataset["instructors"])
    generated_by_student = _count(generated_lessons, "student_id")
    generated_by_car = _count(generated_lessons, "car_id")
    generated_by_instructor = _count(generated_lessons, "instructor_id")
    existing_by_car = _count(existing_lessons, "car_id")
    existing_by_instructor = _count(existing_lessons, "instructor_id")
    lesson_hours = 1.5
    target_lessons = sum(int(student["target_lessons"]) for student in students)
    car_available_hours = sum(_availability_hours(car.get("availability", [])) for car in cars)
    instructor_available_hours = sum(_availability_hours(instructor.get("availability", [])) for instructor in instructors)
    car_busy_hours = sum((generated_by_car.get(car["id"], 0) + existing_by_car.get(car["id"], 0)) * lesson_hours for car in cars)
    instructor_busy_hours = sum(
        (generated_by_instructor.get(instructor["id"], 0) + existing_by_instructor.get(instructor["id"], 0)) * lesson_hours
        for instructor in instructors
    )

    return {
        "request": {
            "request_id": "dataset-reference-schedule",
            "period_start": dataset["metadata"]["period_start"],
            "period_end": dataset["metadata"]["period_end"],
        },
        "summary": {
            "students": len(students),
            "automatic_students": sum(1 for item in students if item["transmission"] == "automatic"),
            "manual_students": sum(1 for item in students if item["transmission"] == "manual"),
            "cars": len(cars),
            "automatic_cars": sum(1 for item in cars if item["transmission"] == "automatic"),
            "manual_cars": sum(1 for item in cars if item["transmission"] == "manual"),
            "instructors": len(instructors),
            "existing_lessons": len(existing_lessons),
            "target_lessons": target_lessons,
            "generated_lessons": len(generated_lessons),
            "unassigned_students": sum(
                1 for student in students if generated_by_student.get(student["id"], 0) < int(student["target_lessons"])
            ),
            "conflicts": 0,
            "best_fitness_score": 0,
            "score": 0,
            "car_available_hours": car_available_hours,
            "car_busy_hours": car_busy_hours,
            "car_idle_hours": max(0.0, car_available_hours - car_busy_hours),
            "instructor_available_hours": instructor_available_hours,
            "instructor_busy_hours": instructor_busy_hours,
            "instructor_idle_hours": max(0.0, instructor_available_hours - instructor_busy_hours),
            "resource_idle_hours": max(0.0, car_available_hours - car_busy_hours) + max(0.0, instructor_available_hours - instructor_busy_hours),
            "all_programs_completed": len(generated_lessons) >= target_lessons,
        },
        "fitness_score_breakdown": {"source": "dataset reference schedule; run planning:dataset:test for Planner score"},
        "score_breakdown": {"source": "dataset reference schedule; run planning:dataset:test for Planner score"},
        "generated_lessons": generated_lessons,
        "cars": [
            {
                "id": car["id"],
                "transmission": car["transmission"],
                "generated_lessons": generated_by_car.get(car["id"], 0),
                "existing_lessons": existing_by_car.get(car["id"], 0),
                "available_hours": _availability_hours(car.get("availability", [])),
                "busy_hours": (generated_by_car.get(car["id"], 0) + existing_by_car.get(car["id"], 0)) * lesson_hours,
                "idle_hours": max(
                    0.0,
                    _availability_hours(car.get("availability", []))
                    - (generated_by_car.get(car["id"], 0) + existing_by_car.get(car["id"], 0)) * lesson_hours,
                ),
            }
            for car in cars
        ],
        "instructors": [
            {
                "id": instructor["id"],
                "generated_lessons": generated_by_instructor.get(instructor["id"], 0),
                "existing_lessons": existing_by_instructor.get(instructor["id"], 0),
                "available_hours": _availability_hours(instructor.get("availability", [])),
                "busy_hours": (
                    generated_by_instructor.get(instructor["id"], 0)
                    + existing_by_instructor.get(instructor["id"], 0)
                )
                * lesson_hours,
                "idle_hours": max(
                    0.0,
                    _availability_hours(instructor.get("availability", []))
                    - (
                        generated_by_instructor.get(instructor["id"], 0)
                        + existing_by_instructor.get(instructor["id"], 0)
                    )
                    * lesson_hours,
                ),
            }
            for instructor in instructors
        ],
        "students": [
            {
                "id": student["id"],
                "transmission": student["transmission"],
                "target_lessons": int(student["target_lessons"]),
                "generated_lessons": generated_by_student.get(student["id"], 0),
                "remaining_lessons": max(0, int(student["target_lessons"]) - generated_by_student.get(student["id"], 0)),
                "last_lesson_at": max(
                    (lesson["start_at"] for lesson in generated_lessons if lesson["student_id"] == student["id"]),
                    default=None,
                ),
                "preferred_weekdays": student["preferences"]["preferred_weekdays"],
                "preferred_time_windows": student["preferences"]["preferred_time_windows"],
            }
            for student in students
        ],
        "lessons_by_date": dict(sorted(_count_by_date(generated_lessons).items())),
        "conflicts": [],
    }


def _count(rows: list[dict[str, Any]], key: str) -> dict[str, int]:
    result: dict[str, int] = {}
    for row in rows:
        result[row[key]] = result.get(row[key], 0) + 1
    return result


def _count_by_date(rows: list[dict[str, Any]]) -> dict[str, int]:
    result: dict[str, int] = {}
    for row in rows:
        day = row["start_at"][:10]
        result[day] = result.get(day, 0) + 1
    return result


def _availability_hours(availability: list[dict[str, Any]]) -> float:
    total = 0.0
    for window in availability:
        if window.get("type") != "working_hours":
            continue
        total += (datetime_fromiso(window["end_at"]) - datetime_fromiso(window["start_at"])).total_seconds() / 3600
    return total


def datetime_fromiso(value: str):
    from datetime import datetime

    return datetime.fromisoformat(value)


def _json_script(payload: dict[str, Any]) -> str:
    return json.dumps(payload, ensure_ascii=False).replace("</", "<\\/")


if __name__ == "__main__":
    main()

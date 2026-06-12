from __future__ import annotations

from collections.abc import Iterator, Sequence
from datetime import datetime, time, timedelta

from .models import LESSON_DURATION, TimeSlot


class TimeSlotCache(Sequence[TimeSlot]):
    def __init__(self, slots: list[TimeSlot]) -> None:
        self._slots = slots
        self._by_index = {slot.index: slot for slot in slots}

    def __getitem__(self, index: int) -> TimeSlot:
        return self._by_index[index]

    def __len__(self) -> int:
        return len(self._slots)

    def __iter__(self) -> Iterator[TimeSlot]:
        return iter(self._slots)

    @property
    def max_index(self) -> int:
        return len(self._slots) - 1


def build_time_slot_cache(period_start, period_end) -> TimeSlotCache:
    slots: list[TimeSlot] = []
    cursor = period_start
    while cursor + LESSON_DURATION <= period_end:
        slots.append(TimeSlot(index=len(slots), start_at=cursor, end_at=cursor + LESSON_DURATION))
        cursor += LESSON_DURATION
    return TimeSlotCache(slots)


def build_daily_time_slot_cache(period_start, period_end, available_from: time, available_to: time) -> TimeSlotCache:
    slots: list[TimeSlot] = []
    current_date = period_start.date()
    end_date = period_end.date()
    while current_date <= end_date:
        cursor = datetime.combine(current_date, available_from, tzinfo=period_start.tzinfo)
        day_end = datetime.combine(current_date, available_to, tzinfo=period_start.tzinfo)
        while cursor + LESSON_DURATION <= day_end:
            slot_end = cursor + LESSON_DURATION
            if cursor >= period_start and slot_end <= period_end:
                slots.append(TimeSlot(index=len(slots), start_at=cursor, end_at=slot_end))
            cursor += LESSON_DURATION
        current_date += timedelta(days=1)
    return TimeSlotCache(slots)

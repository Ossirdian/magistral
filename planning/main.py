from app.kafka_consumer import PlanningWorker
from app.settings import Settings


def main() -> None:
    settings = Settings.from_env()
    worker = PlanningWorker(settings)
    worker.run_forever()


if __name__ == "__main__":
    main()

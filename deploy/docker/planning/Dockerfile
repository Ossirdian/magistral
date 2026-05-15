FROM python:3.12-slim

WORKDIR /app

COPY planning/requirements.txt ./planning/requirements.txt
RUN pip install --no-cache-dir -r planning/requirements.txt

COPY planning ./planning

WORKDIR /app/planning

ENTRYPOINT ["python", "main.py"]

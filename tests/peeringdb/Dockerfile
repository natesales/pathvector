FROM python:3.12-slim

WORKDIR /app

RUN pip install --no-cache-dir flask

COPY . .

EXPOSE 5000

CMD ["python", "peeringdb-test-api.py"]

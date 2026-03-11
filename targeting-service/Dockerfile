FROM python:3.9-slim

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

WORKDIR /targeting-service


RUN apt-get update \
    && apt-get install -y --no-install-recommends postgresql-client \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .
RUN pip uninstall -y werkzeug flask psycopg && \ 
    pip install --no-cache-dir -r requirements.txt


COPY app.py .

EXPOSE 8003

RUN pip uninstall -y werkzeug flask psycopg || true && \
    pip install --no-cache-dir --user -r requirements.txt && \
    pip show werkzeug && pip show flask

CMD ["gunicorn", "--bind", "0.0.0.0:8003", "app:app"]
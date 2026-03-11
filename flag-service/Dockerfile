FROM python:3.11-slim AS builder

WORKDIR /app

COPY requirements.txt .
RUN pip uninstall -y werkzeug flask psycopg && \ 
    pip install --user -r requirements.txt

FROM python:3.11-slim

WORKDIR /app

COPY --from=builder /root/.local /root/.local

ENV PATH=/root/.local/bin:$PATH

COPY . .

# Garante que não há versões conflitantes no ambiente final
RUN pip uninstall -y werkzeug flask psycopg || true && \
    pip install --no-cache-dir --user -r requirements.txt && \
    pip show werkzeug && pip show flask

CMD ["python", "app.py"]

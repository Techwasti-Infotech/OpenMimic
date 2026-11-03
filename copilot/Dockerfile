FROM python:3.12-slim

WORKDIR /app
COPY copilot/ /app/copilot/
RUN pip install --no-cache-dir fastapi uvicorn pydantic protobuf

EXPOSE 8000
CMD ["uvicorn", "copilot.main:app", "--host", "0.0.0.0", "--port", "8000"]


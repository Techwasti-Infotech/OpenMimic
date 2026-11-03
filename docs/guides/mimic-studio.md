# Mimic Studio Manual

**Mimic Studio** is the visual IDE for OpenMimic, built with React 18, Vite, and Mantine. It enables interactive design of data generation streams, live previews, and real-time observability.

---

## 🎨 Studio Layout

Mimic Studio provides three core tabs in the left sidebar:

1. **Generator Designer:** Build, edit, and configure data schemas and sinks.
2. **Mimic Copilot:** Interact with the natural-language generation assistant.
3. **Observability Dashboard:** Monitor throughput, p99 latencies, and error rates in real time.

---

## 🛠️ Step-by-Step Workflow

### 1. Designing a Generator

1. Navigate to **Generator Designer**.
2. Specify the **Generator Name**, **Seed**, and **Event Limit**.
3. Set your target destination (e.g. Kafka topic `user.signups`).
4. Click **Add Field** to introduce new properties to the generated JSON records.
5. Choose a generator type:
   - `uuid`: Unique identifier.
   - `distribution`: Normal or Uniform distribution for quantitative variables (e.g. latency, risk scores, payment amounts).
   - `now`: Logical timestamp formatted as RFC3339 or epoch.
   - `string_faker`: Contextual strings (emails, names, locations).
   - `weighted_choice`: Probabilistic branches (success vs failure).

---

### 2. Live Preview Streaming

Before streaming live events into production message queues or databases, use the **Preview (10)** feature:
- Clicking **Preview (10)** triggers an isolated gRPC request to `StreamPreview`.
- The engine executes 10 iterations in-memory using the exact logical clock and seed configured.
- The output displays in real time on the right-hand panel as formatted JSON records.

---

### 3. Running & Stopping Pipelines

- Click **Run Generator** to dispatch the configuration to the OpenMimic Engine via `POST /api/generators/start`.
- The status pill in the header transitions to `Engine: RUNNING`.
- To halt stream execution at any time, click **Stop Generator**.

---

### 4. Real-Time Observability

Switch to the **Observability Dashboard** tab during execution:
- View live **Throughput (events/sec)** tracked with smooth Recharts visualizations.
- Track total events produced and error count.
- Monitor execution latencies.


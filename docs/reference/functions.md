# Functions & Generator Reference

OpenMimic provides a comprehensive suite of declarative function generators. Each field in `mimic.json` specifies one of the following generator types.

---

## 1. `uuid`

Generates standard RFC 4122 v4 UUIDs deterministically based on the generator seed.

### Schema
```json
{
  "uuid": {
    "uppercase": false
  }
}
```

### Options
- `uppercase` *(boolean, default: `false`)*: If `true`, returns the UUID in uppercase (e.g. `3F982A7E-...`).

---

## 2. `distribution`

Outputs mathematical distributions (Gaussian / Normal or Uniform).

### Normal (Gaussian) Distribution
Generates values around a mean with standard deviation:
```json
{
  "distribution": {
    "normal": {
      "mean": 100.0,
      "std_dev": 15.0
    }
  }
}
```

### Uniform Distribution
Generates values evenly spread between `min` and `max`:
```json
{
  "distribution": {
    "uniform": {
      "min": 10.0,
      "max": 50.0
    }
  }
}
```

---

## 3. `now`

Generates deterministic timestamps that advance with the generator logical clock.

### Schema
```json
{
  "now": {
    "format": "RFC3339"
  }
}
```

### Options
- `format` *(string)*:
  - `"RFC3339"`: `2026-06-15T10:00:00Z`
  - `"epoch_ms"`: Unix timestamp in milliseconds (`1750000000000`)
  - `"epoch_s"`: Unix timestamp in seconds (`1750000000`)
  - Custom Go layout string (e.g. `"2006-01-02 15:04:05"`).

---

## 4. `sequential_integer`

Generates monotonically incrementing or stepping integers.

### Schema
```json
{
  "sequential_integer": {
    "start": 1,
    "step": 1
  }
}
```

### Output Formula
$$\text{value} = \text{start} + (\text{event\_index} \times \text{step})$$

---

## 5. `string_faker`

Synthesizes context-aware strings and pattern-based identifiers.

### Semantic Categories
```json
{
  "string_faker": {
    "category": "email"
  }
}
```
Available categories:
- `"name"` / `"full_name"`: Generates realistic human names (e.g. `"Alice Smith"`).
- `"first_name"`, `"last_name"`.
- `"email"`: Generates formatted email addresses.
- `"city"`: Major world cities.
- `"country"`: Global country names.

### Custom Mask / Pattern
You can specify a pattern using `#` for random digits:
```json
{
  "string_faker": {
    "pattern": "ORDER_####"
  }
}
```
Output: `ORDER_4821`, `ORDER_0945`.

---

## 6. `weighted_choice`

Simulates probabilistic distributions, failure rates, or categorical states.

### Schema
```json
{
  "weighted_choice": {
    "options": [
      { "value": "SUCCESS", "weight": 95.0 },
      { "value": "FAILURE", "weight": 5.0 }
    ]
  }
}
```
In this example, `"FAILURE"` is produced with an exact 5% probability across the event stream.

---

## 7. `constant`

Outputs a fixed literal value.

```json
{
  "constant": {
    "string_val": "ACTIVE"
  }
}
```
Supports `string_val`, `int_val`, `double_val`, and `bool_val`.

---

## 8. `binary_media` (Synthetic Audio & Images)

Generates valid binary media containers directly into streaming payloads (useful for multimodal pipelines, computer vision benchmarks, and media processing ingestion).

### Schema
```json
{
  "binary_media": {
    "format": "wav",
    "sample_rate": 44100,
    "duration_seconds": 1.0,
    "frequency_hz": 440.0
  }
}
```

### Supported Media Formats:
- **`wav`**: Generates RFC PCM WAV audio streams containing a pure sinusoidal tone (e.g. 440 Hz standard A4) with valid `RIFF` and `WAVEfmt ` headers.
- **`png`**: Generates valid PNG images with official magic header `0x89 0x50 0x4E 0x47 0x0D 0x0A 0x1A 0x0A`, `IHDR`, and `IEND` chunks.
- **`jpeg`**: Generates valid JPEG images with SOI marker `0xFF 0xD8`, JFIF application header, and EOI marker `0xFF 0xD9`.

---

## 9. `aggregation` (Streaming CEP Window Rollups)

Calculates streaming rollups (`sum`, `avg`, `min`, `max`, `count`) over sliding or tumbling event windows from correlated upstream streams.

### Schema
```json
{
  "aggregation": {
    "source_stream_id": "stream-orders",
    "window_type": "sliding",
    "window_size_events": 50,
    "aggregation_op": "avg",
    "target_field": "order_amount"
  }
}
```

---

## 10. `markov_chain` & `entity_pool`

- **`entity_pool`**: Selects foreign keys from a pre-allocated bounded pool of entities (e.g. `cust_00001` through `cust_00500`) to maintain relational referential integrity across topics.
- **`markov_chain`**: Simulates realistic state progressions (e.g. `CART -> CHECKOUT -> PAYMENT_PENDING -> ORDER_COMPLETED`) according to transition probability matrices.



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


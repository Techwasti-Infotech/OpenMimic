import json
import re
from typing import Dict, Any, Optional
from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from google.protobuf import json_format

# Import generated protobuf stubs
from copilot.proto import mimic_config_pb2

app = FastAPI(title="OpenMimic Copilot Service", version="1.0.0")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

class GenerateConfigRequest(BaseModel):
    prompt: str

class GenerateConfigResponse(BaseModel):
    config: Dict[str, Any]
    raw_prompt: str
    validated: bool
    pii_warnings: Optional[list[str]] = None

def scan_pii_fields(fields: list[Dict[str, Any]]) -> list[str]:
    """
    Phase 8: Inspects generated configuration fields and flags unmasked PII.
    """
    pii_indicators = ["email", "ssn", "phone", "credit_card", "password", "ip_address", "passport", "full_name"]
    warnings = []

    for f in fields:
        name = f.get("name", "").lower()
        for pii in pii_indicators:
            if pii in name:
                gen = f.get("generator", {})
                # Flag if raw string without masking
                warnings.append(
                    f"Field '{f.get('name')}' appears to be PII ({pii}). Suggest applying deterministic hashing or token masking."
                )
                break
    return warnings

def parse_natural_language_to_config(prompt: str) -> Dict[str, Any]:
    """
    Translates natural language prompts into valid OpenMimic GeneratorConfig schemas.
    Includes rules for user signups, failure rates, Kafka destinations, distributions, etc.
    """
    p = prompt.lower()

    # Detect event limit (e.g. "100 user signup events" or "1000 events")
    limit_match = re.search(r'(\d+)\s+(?:events?|user|signup|records?)', p)
    limit = int(limit_match.group(1)) if limit_match else 100

    # Detect failure rate (e.g. "5% failure rate")
    fail_match = re.search(r'(\d+(?:\.\d+)?)\s*%\s*fail', p)
    fail_rate = float(fail_match.group(1)) if fail_match else 5.0
    success_rate = max(0.0, 100.0 - fail_rate)

    fields = [
        {
            "name": "user_id",
            "generator": {
                "uuid": {"uppercase": False}
            }
        },
        {
            "name": "email",
            "generator": {
                "string_faker": {"category": "email"}
            }
        },
        {
            "name": "status",
            "generator": {
                "weighted_choice": {
                    "options": [
                        {"value": "SUCCESS", "weight": success_rate},
                        {"value": "FAILURE", "weight": fail_rate}
                    ]
                }
            }
        },
        {
            "name": "signup_timestamp",
            "generator": {
                "now": {"format": "RFC3339"}
            }
        }
    ]

    # If prompt mentions latency, amount, or response time
    if "latency" in p or "duration" in p or "time" in p:
        fields.append({
            "name": "response_time_ms",
            "generator": {
                "distribution": {
                    "normal": {"mean": 120.0, "std_dev": 25.0}
                }
            }
        })

    # Detect traffic pattern requests (bursts, diurnal, poisson)
    traffic_profile = None
    if "burst" in p or "spike" in p:
        traffic_profile = {
            "burst": {
                "normal_rate": 10,
                "burst_rate": 100,
                "interval_seconds": 20,
                "duration_seconds": 5
            }
        }
    elif "poisson" in p or "stochastic" in p:
        traffic_profile = {
            "poisson": {
                "lambda_rate": 25.0
            }
        }
    elif "day" in p or "diurnal" in p or "cycle" in p:
        traffic_profile = {
            "diurnal": {
                "base_rate": 5,
                "peak_rate": 50,
                "period_seconds": 60
            }
        }

    # Detect entity pool / customer pool references
    if "pool" in p or "repeat user" in p or "returning customer" in p:
        fields[0] = {
            "name": "user_id",
            "generator": {
                "entity_pool": {
                    "pool_name": "active_users",
                    "pool_size": 100,
                    "entity_prefix": "usr_"
                }
            }
        }

    # Detect lifecycle / state machine requests
    if "state machine" in p or "funnel" in p or "journey" in p:
        fields[2] = {
            "name": "journey_state",
            "generator": {
                "markov_chain": {
                    "initial_state": "LANDING_PAGE",
                    "states": [
                        {
                            "state_name": "LANDING_PAGE",
                            "transitions": [
                                {"to_state": "PRODUCT_VIEW", "probability": 0.8},
                                {"to_state": "BOUNCE", "probability": 0.2}
                            ]
                        },
                        {
                            "state_name": "PRODUCT_VIEW",
                            "transitions": [
                                {"to_state": "CART_ADD", "probability": 0.6},
                                {"to_state": "EXIT", "probability": 0.4}
                            ]
                        },
                        {
                            "state_name": "CART_ADD",
                            "transitions": [
                                {"to_state": "CHECKOUT", "probability": 0.7},
                                {"to_state": "ABANDON", "probability": 0.3}
                            ]
                        }
                    ]
                }
            }
        }

    config = {
        "id": "gen_copilot_signup",
        "name": "Copilot Generated Signups",
        "seed": 42,
        "event_limit": limit,
        "rate_limit_per_second": 20,
        "sink": {
            "kafka": {
                "brokers": ["localhost:9092"],
                "topic": "user.signups",
                "client_id": "openmimic-copilot"
            }
        },
        "fields": fields
    }

    if traffic_profile:
        config["traffic_profile"] = traffic_profile

    return config

def validate_against_proto(config_dict: Dict[str, Any]) -> mimic_config_pb2.GeneratorConfig:
    """
    Validates a dictionary against the official mimic_config.proto schema.
    """
    msg = mimic_config_pb2.GeneratorConfig()
    json_format.ParseDict(config_dict, msg)
    return msg

@app.post("/generate-config", response_model=GenerateConfigResponse)
def generate_config(req: GenerateConfigRequest):
    if not req.prompt.strip():
        raise HTTPException(status_code=400, detail="Prompt cannot be empty")

    config_dict = parse_natural_language_to_config(req.prompt)

    try:
        # Validate against proto schema
        validate_against_proto(config_dict)
    except Exception as e:
        raise HTTPException(status_code=422, detail=f"Generated config failed proto validation: {str(e)}")

    # Phase 8: PII Scanning
    pii_warnings = scan_pii_fields(config_dict.get("fields", []))

    return GenerateConfigResponse(
        config=config_dict,
        raw_prompt=req.prompt,
        validated=True,
        pii_warnings=pii_warnings
    )

@app.get("/health")
def health():
    return {"status": "ok"}


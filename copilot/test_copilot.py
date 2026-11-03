import unittest
from copilot.main import parse_natural_language_to_config, validate_against_proto
from copilot.proto import mimic_config_pb2

class TestCopilotService(unittest.TestCase):
    def test_signup_prompt_with_failure_rate(self):
        prompt = "Generate a stream of 100 user signup events with a 5% failure rate"
        cfg = parse_natural_language_to_config(prompt)

        # 1. Assert proto schema validation passes without exception
        proto_msg = validate_against_proto(cfg)
        self.assertIsInstance(proto_msg, mimic_config_pb2.GeneratorConfig)

        # 2. Check event limit
        self.assertEqual(proto_msg.event_limit, 100)

        # 3. Check Kafka sink
        self.assertTrue(proto_msg.sink.HasField("kafka"))
        self.assertEqual(proto_msg.sink.kafka.topic, "user.signups")

        # 4. Check fields and failure rate state configuration
        field_names = [f.name for f in proto_msg.fields]
        self.assertIn("user_id", field_names)
        self.assertIn("status", field_names)

        # Find status field
        status_field = next(f for f in proto_msg.fields if f.name == "status")
        self.assertTrue(status_field.generator.HasField("weighted_choice"))
        options = status_field.generator.weighted_choice.options
        opt_map = {opt.value: opt.weight for opt in options}

        self.assertEqual(opt_map.get("FAILURE"), 5.0)
        self.assertEqual(opt_map.get("SUCCESS"), 95.0)

    def test_deterministic_seed_present(self):
        cfg = parse_natural_language_to_config("test random stream")
        self.assertGreater(cfg["seed"], 0)

if __name__ == "__main__":
    unittest.main()


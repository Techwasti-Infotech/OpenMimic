from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class GeneratorConfig(_message.Message):
    __slots__ = ("id", "name", "seed", "event_limit", "rate_limit_per_second", "sink", "fields", "traffic_profile", "constraints", "playbooks")
    ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    SEED_FIELD_NUMBER: _ClassVar[int]
    EVENT_LIMIT_FIELD_NUMBER: _ClassVar[int]
    RATE_LIMIT_PER_SECOND_FIELD_NUMBER: _ClassVar[int]
    SINK_FIELD_NUMBER: _ClassVar[int]
    FIELDS_FIELD_NUMBER: _ClassVar[int]
    TRAFFIC_PROFILE_FIELD_NUMBER: _ClassVar[int]
    CONSTRAINTS_FIELD_NUMBER: _ClassVar[int]
    PLAYBOOKS_FIELD_NUMBER: _ClassVar[int]
    id: str
    name: str
    seed: int
    event_limit: int
    rate_limit_per_second: int
    sink: SinkConfig
    fields: _containers.RepeatedCompositeFieldContainer[Field]
    traffic_profile: TrafficProfile
    constraints: _containers.RepeatedCompositeFieldContainer[Constraint]
    playbooks: _containers.RepeatedCompositeFieldContainer[AnomalyPlaybook]
    def __init__(self, id: _Optional[str] = ..., name: _Optional[str] = ..., seed: _Optional[int] = ..., event_limit: _Optional[int] = ..., rate_limit_per_second: _Optional[int] = ..., sink: _Optional[_Union[SinkConfig, _Mapping]] = ..., fields: _Optional[_Iterable[_Union[Field, _Mapping]]] = ..., traffic_profile: _Optional[_Union[TrafficProfile, _Mapping]] = ..., constraints: _Optional[_Iterable[_Union[Constraint, _Mapping]]] = ..., playbooks: _Optional[_Iterable[_Union[AnomalyPlaybook, _Mapping]]] = ...) -> None: ...

class Constraint(_message.Message):
    __slots__ = ("id", "name", "expression", "max_resample_retries")
    ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    EXPRESSION_FIELD_NUMBER: _ClassVar[int]
    MAX_RESAMPLE_RETRIES_FIELD_NUMBER: _ClassVar[int]
    id: str
    name: str
    expression: str
    max_resample_retries: int
    def __init__(self, id: _Optional[str] = ..., name: _Optional[str] = ..., expression: _Optional[str] = ..., max_resample_retries: _Optional[int] = ...) -> None: ...

class AnomalyPlaybook(_message.Message):
    __slots__ = ("name", "anomaly_rate", "injection_type")
    NAME_FIELD_NUMBER: _ClassVar[int]
    ANOMALY_RATE_FIELD_NUMBER: _ClassVar[int]
    INJECTION_TYPE_FIELD_NUMBER: _ClassVar[int]
    name: str
    anomaly_rate: float
    injection_type: str
    def __init__(self, name: _Optional[str] = ..., anomaly_rate: _Optional[float] = ..., injection_type: _Optional[str] = ...) -> None: ...

class Workspace(_message.Message):
    __slots__ = ("id", "name", "owner_id", "member_ids")
    ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    OWNER_ID_FIELD_NUMBER: _ClassVar[int]
    MEMBER_IDS_FIELD_NUMBER: _ClassVar[int]
    id: str
    name: str
    owner_id: str
    member_ids: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, id: _Optional[str] = ..., name: _Optional[str] = ..., owner_id: _Optional[str] = ..., member_ids: _Optional[_Iterable[str]] = ...) -> None: ...

class Template(_message.Message):
    __slots__ = ("id", "title", "category", "config")
    ID_FIELD_NUMBER: _ClassVar[int]
    TITLE_FIELD_NUMBER: _ClassVar[int]
    CATEGORY_FIELD_NUMBER: _ClassVar[int]
    CONFIG_FIELD_NUMBER: _ClassVar[int]
    id: str
    title: str
    category: str
    config: GeneratorConfig
    def __init__(self, id: _Optional[str] = ..., title: _Optional[str] = ..., category: _Optional[str] = ..., config: _Optional[_Union[GeneratorConfig, _Mapping]] = ...) -> None: ...

class AuditEvent(_message.Message):
    __slots__ = ("id", "user_id", "action", "generator_id", "timestamp_ms", "details")
    ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    ACTION_FIELD_NUMBER: _ClassVar[int]
    GENERATOR_ID_FIELD_NUMBER: _ClassVar[int]
    TIMESTAMP_MS_FIELD_NUMBER: _ClassVar[int]
    DETAILS_FIELD_NUMBER: _ClassVar[int]
    id: str
    user_id: str
    action: str
    generator_id: str
    timestamp_ms: int
    details: str
    def __init__(self, id: _Optional[str] = ..., user_id: _Optional[str] = ..., action: _Optional[str] = ..., generator_id: _Optional[str] = ..., timestamp_ms: _Optional[int] = ..., details: _Optional[str] = ...) -> None: ...

class TrafficProfile(_message.Message):
    __slots__ = ("constant", "diurnal", "poisson", "burst")
    CONSTANT_FIELD_NUMBER: _ClassVar[int]
    DIURNAL_FIELD_NUMBER: _ClassVar[int]
    POISSON_FIELD_NUMBER: _ClassVar[int]
    BURST_FIELD_NUMBER: _ClassVar[int]
    constant: ConstantRate
    diurnal: DiurnalRate
    poisson: PoissonArrival
    burst: BurstSpikeRate
    def __init__(self, constant: _Optional[_Union[ConstantRate, _Mapping]] = ..., diurnal: _Optional[_Union[DiurnalRate, _Mapping]] = ..., poisson: _Optional[_Union[PoissonArrival, _Mapping]] = ..., burst: _Optional[_Union[BurstSpikeRate, _Mapping]] = ...) -> None: ...

class ConstantRate(_message.Message):
    __slots__ = ("rate_per_second",)
    RATE_PER_SECOND_FIELD_NUMBER: _ClassVar[int]
    rate_per_second: int
    def __init__(self, rate_per_second: _Optional[int] = ...) -> None: ...

class DiurnalRate(_message.Message):
    __slots__ = ("base_rate", "peak_rate", "period_seconds")
    BASE_RATE_FIELD_NUMBER: _ClassVar[int]
    PEAK_RATE_FIELD_NUMBER: _ClassVar[int]
    PERIOD_SECONDS_FIELD_NUMBER: _ClassVar[int]
    base_rate: int
    peak_rate: int
    period_seconds: int
    def __init__(self, base_rate: _Optional[int] = ..., peak_rate: _Optional[int] = ..., period_seconds: _Optional[int] = ...) -> None: ...

class PoissonArrival(_message.Message):
    __slots__ = ("lambda_rate",)
    LAMBDA_RATE_FIELD_NUMBER: _ClassVar[int]
    lambda_rate: float
    def __init__(self, lambda_rate: _Optional[float] = ...) -> None: ...

class BurstSpikeRate(_message.Message):
    __slots__ = ("normal_rate", "burst_rate", "interval_seconds", "duration_seconds")
    NORMAL_RATE_FIELD_NUMBER: _ClassVar[int]
    BURST_RATE_FIELD_NUMBER: _ClassVar[int]
    INTERVAL_SECONDS_FIELD_NUMBER: _ClassVar[int]
    DURATION_SECONDS_FIELD_NUMBER: _ClassVar[int]
    normal_rate: int
    burst_rate: int
    interval_seconds: int
    duration_seconds: int
    def __init__(self, normal_rate: _Optional[int] = ..., burst_rate: _Optional[int] = ..., interval_seconds: _Optional[int] = ..., duration_seconds: _Optional[int] = ...) -> None: ...

class SinkConfig(_message.Message):
    __slots__ = ("kafka", "postgres", "s3", "stdout")
    KAFKA_FIELD_NUMBER: _ClassVar[int]
    POSTGRES_FIELD_NUMBER: _ClassVar[int]
    S3_FIELD_NUMBER: _ClassVar[int]
    STDOUT_FIELD_NUMBER: _ClassVar[int]
    kafka: KafkaSink
    postgres: PostgresSink
    s3: S3Sink
    stdout: StdoutSink
    def __init__(self, kafka: _Optional[_Union[KafkaSink, _Mapping]] = ..., postgres: _Optional[_Union[PostgresSink, _Mapping]] = ..., s3: _Optional[_Union[S3Sink, _Mapping]] = ..., stdout: _Optional[_Union[StdoutSink, _Mapping]] = ...) -> None: ...

class KafkaSink(_message.Message):
    __slots__ = ("brokers", "topic", "client_id")
    BROKERS_FIELD_NUMBER: _ClassVar[int]
    TOPIC_FIELD_NUMBER: _ClassVar[int]
    CLIENT_ID_FIELD_NUMBER: _ClassVar[int]
    brokers: _containers.RepeatedScalarFieldContainer[str]
    topic: str
    client_id: str
    def __init__(self, brokers: _Optional[_Iterable[str]] = ..., topic: _Optional[str] = ..., client_id: _Optional[str] = ...) -> None: ...

class PostgresSink(_message.Message):
    __slots__ = ("connection_string", "table", "columns")
    CONNECTION_STRING_FIELD_NUMBER: _ClassVar[int]
    TABLE_FIELD_NUMBER: _ClassVar[int]
    COLUMNS_FIELD_NUMBER: _ClassVar[int]
    connection_string: str
    table: str
    columns: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, connection_string: _Optional[str] = ..., table: _Optional[str] = ..., columns: _Optional[_Iterable[str]] = ...) -> None: ...

class S3Sink(_message.Message):
    __slots__ = ("bucket", "region", "prefix", "file_format")
    BUCKET_FIELD_NUMBER: _ClassVar[int]
    REGION_FIELD_NUMBER: _ClassVar[int]
    PREFIX_FIELD_NUMBER: _ClassVar[int]
    FILE_FORMAT_FIELD_NUMBER: _ClassVar[int]
    bucket: str
    region: str
    prefix: str
    file_format: str
    def __init__(self, bucket: _Optional[str] = ..., region: _Optional[str] = ..., prefix: _Optional[str] = ..., file_format: _Optional[str] = ...) -> None: ...

class StdoutSink(_message.Message):
    __slots__ = ("pretty",)
    PRETTY_FIELD_NUMBER: _ClassVar[int]
    pretty: bool
    def __init__(self, pretty: _Optional[bool] = ...) -> None: ...

class Field(_message.Message):
    __slots__ = ("name", "generator")
    NAME_FIELD_NUMBER: _ClassVar[int]
    GENERATOR_FIELD_NUMBER: _ClassVar[int]
    name: str
    generator: FunctionGenerator
    def __init__(self, name: _Optional[str] = ..., generator: _Optional[_Union[FunctionGenerator, _Mapping]] = ...) -> None: ...

class FunctionGenerator(_message.Message):
    __slots__ = ("uuid", "distribution", "now", "sequential_integer", "string_faker", "constant", "weighted_choice", "markov_chain", "entity_pool")
    UUID_FIELD_NUMBER: _ClassVar[int]
    DISTRIBUTION_FIELD_NUMBER: _ClassVar[int]
    NOW_FIELD_NUMBER: _ClassVar[int]
    SEQUENTIAL_INTEGER_FIELD_NUMBER: _ClassVar[int]
    STRING_FAKER_FIELD_NUMBER: _ClassVar[int]
    CONSTANT_FIELD_NUMBER: _ClassVar[int]
    WEIGHTED_CHOICE_FIELD_NUMBER: _ClassVar[int]
    MARKOV_CHAIN_FIELD_NUMBER: _ClassVar[int]
    ENTITY_POOL_FIELD_NUMBER: _ClassVar[int]
    uuid: UuidFunction
    distribution: DistributionFunction
    now: NowFunction
    sequential_integer: SequentialIntegerFunction
    string_faker: StringFakerFunction
    constant: ConstantFunction
    weighted_choice: WeightedChoiceFunction
    markov_chain: MarkovChainFunction
    entity_pool: EntityPoolFunction
    def __init__(self, uuid: _Optional[_Union[UuidFunction, _Mapping]] = ..., distribution: _Optional[_Union[DistributionFunction, _Mapping]] = ..., now: _Optional[_Union[NowFunction, _Mapping]] = ..., sequential_integer: _Optional[_Union[SequentialIntegerFunction, _Mapping]] = ..., string_faker: _Optional[_Union[StringFakerFunction, _Mapping]] = ..., constant: _Optional[_Union[ConstantFunction, _Mapping]] = ..., weighted_choice: _Optional[_Union[WeightedChoiceFunction, _Mapping]] = ..., markov_chain: _Optional[_Union[MarkovChainFunction, _Mapping]] = ..., entity_pool: _Optional[_Union[EntityPoolFunction, _Mapping]] = ...) -> None: ...

class MarkovTransition(_message.Message):
    __slots__ = ("to_state", "probability")
    TO_STATE_FIELD_NUMBER: _ClassVar[int]
    PROBABILITY_FIELD_NUMBER: _ClassVar[int]
    to_state: str
    probability: float
    def __init__(self, to_state: _Optional[str] = ..., probability: _Optional[float] = ...) -> None: ...

class MarkovState(_message.Message):
    __slots__ = ("state_name", "transitions")
    STATE_NAME_FIELD_NUMBER: _ClassVar[int]
    TRANSITIONS_FIELD_NUMBER: _ClassVar[int]
    state_name: str
    transitions: _containers.RepeatedCompositeFieldContainer[MarkovTransition]
    def __init__(self, state_name: _Optional[str] = ..., transitions: _Optional[_Iterable[_Union[MarkovTransition, _Mapping]]] = ...) -> None: ...

class MarkovChainFunction(_message.Message):
    __slots__ = ("initial_state", "states")
    INITIAL_STATE_FIELD_NUMBER: _ClassVar[int]
    STATES_FIELD_NUMBER: _ClassVar[int]
    initial_state: str
    states: _containers.RepeatedCompositeFieldContainer[MarkovState]
    def __init__(self, initial_state: _Optional[str] = ..., states: _Optional[_Iterable[_Union[MarkovState, _Mapping]]] = ...) -> None: ...

class EntityPoolFunction(_message.Message):
    __slots__ = ("pool_name", "pool_size", "entity_prefix")
    POOL_NAME_FIELD_NUMBER: _ClassVar[int]
    POOL_SIZE_FIELD_NUMBER: _ClassVar[int]
    ENTITY_PREFIX_FIELD_NUMBER: _ClassVar[int]
    pool_name: str
    pool_size: int
    entity_prefix: str
    def __init__(self, pool_name: _Optional[str] = ..., pool_size: _Optional[int] = ..., entity_prefix: _Optional[str] = ...) -> None: ...

class UuidFunction(_message.Message):
    __slots__ = ("uppercase",)
    UPPERCASE_FIELD_NUMBER: _ClassVar[int]
    uppercase: bool
    def __init__(self, uppercase: _Optional[bool] = ...) -> None: ...

class DistributionFunction(_message.Message):
    __slots__ = ("normal", "uniform")
    NORMAL_FIELD_NUMBER: _ClassVar[int]
    UNIFORM_FIELD_NUMBER: _ClassVar[int]
    normal: NormalDistribution
    uniform: UniformDistribution
    def __init__(self, normal: _Optional[_Union[NormalDistribution, _Mapping]] = ..., uniform: _Optional[_Union[UniformDistribution, _Mapping]] = ...) -> None: ...

class NormalDistribution(_message.Message):
    __slots__ = ("mean", "std_dev")
    MEAN_FIELD_NUMBER: _ClassVar[int]
    STD_DEV_FIELD_NUMBER: _ClassVar[int]
    mean: float
    std_dev: float
    def __init__(self, mean: _Optional[float] = ..., std_dev: _Optional[float] = ...) -> None: ...

class UniformDistribution(_message.Message):
    __slots__ = ("min", "max")
    MIN_FIELD_NUMBER: _ClassVar[int]
    MAX_FIELD_NUMBER: _ClassVar[int]
    min: float
    max: float
    def __init__(self, min: _Optional[float] = ..., max: _Optional[float] = ...) -> None: ...

class NowFunction(_message.Message):
    __slots__ = ("format",)
    FORMAT_FIELD_NUMBER: _ClassVar[int]
    format: str
    def __init__(self, format: _Optional[str] = ...) -> None: ...

class SequentialIntegerFunction(_message.Message):
    __slots__ = ("start", "step")
    START_FIELD_NUMBER: _ClassVar[int]
    STEP_FIELD_NUMBER: _ClassVar[int]
    start: int
    step: int
    def __init__(self, start: _Optional[int] = ..., step: _Optional[int] = ...) -> None: ...

class StringFakerFunction(_message.Message):
    __slots__ = ("category", "pattern")
    CATEGORY_FIELD_NUMBER: _ClassVar[int]
    PATTERN_FIELD_NUMBER: _ClassVar[int]
    category: str
    pattern: str
    def __init__(self, category: _Optional[str] = ..., pattern: _Optional[str] = ...) -> None: ...

class ConstantFunction(_message.Message):
    __slots__ = ("string_val", "int_val", "double_val", "bool_val")
    STRING_VAL_FIELD_NUMBER: _ClassVar[int]
    INT_VAL_FIELD_NUMBER: _ClassVar[int]
    DOUBLE_VAL_FIELD_NUMBER: _ClassVar[int]
    BOOL_VAL_FIELD_NUMBER: _ClassVar[int]
    string_val: str
    int_val: int
    double_val: float
    bool_val: bool
    def __init__(self, string_val: _Optional[str] = ..., int_val: _Optional[int] = ..., double_val: _Optional[float] = ..., bool_val: _Optional[bool] = ...) -> None: ...

class WeightedChoiceOption(_message.Message):
    __slots__ = ("value", "weight")
    VALUE_FIELD_NUMBER: _ClassVar[int]
    WEIGHT_FIELD_NUMBER: _ClassVar[int]
    value: str
    weight: float
    def __init__(self, value: _Optional[str] = ..., weight: _Optional[float] = ...) -> None: ...

class WeightedChoiceFunction(_message.Message):
    __slots__ = ("options",)
    OPTIONS_FIELD_NUMBER: _ClassVar[int]
    options: _containers.RepeatedCompositeFieldContainer[WeightedChoiceOption]
    def __init__(self, options: _Optional[_Iterable[_Union[WeightedChoiceOption, _Mapping]]] = ...) -> None: ...

class StartGeneratorRequest(_message.Message):
    __slots__ = ("config",)
    CONFIG_FIELD_NUMBER: _ClassVar[int]
    config: GeneratorConfig
    def __init__(self, config: _Optional[_Union[GeneratorConfig, _Mapping]] = ...) -> None: ...

class StartGeneratorResponse(_message.Message):
    __slots__ = ("generator_id", "status")
    GENERATOR_ID_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    generator_id: str
    status: str
    def __init__(self, generator_id: _Optional[str] = ..., status: _Optional[str] = ...) -> None: ...

class StopGeneratorRequest(_message.Message):
    __slots__ = ("generator_id",)
    GENERATOR_ID_FIELD_NUMBER: _ClassVar[int]
    generator_id: str
    def __init__(self, generator_id: _Optional[str] = ...) -> None: ...

class StopGeneratorResponse(_message.Message):
    __slots__ = ("generator_id", "status")
    GENERATOR_ID_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    generator_id: str
    status: str
    def __init__(self, generator_id: _Optional[str] = ..., status: _Optional[str] = ...) -> None: ...

class StreamPreviewRequest(_message.Message):
    __slots__ = ("config", "preview_count")
    CONFIG_FIELD_NUMBER: _ClassVar[int]
    PREVIEW_COUNT_FIELD_NUMBER: _ClassVar[int]
    config: GeneratorConfig
    preview_count: int
    def __init__(self, config: _Optional[_Union[GeneratorConfig, _Mapping]] = ..., preview_count: _Optional[int] = ...) -> None: ...

class StreamPreviewResponse(_message.Message):
    __slots__ = ("index", "json_payload", "timestamp_ms")
    INDEX_FIELD_NUMBER: _ClassVar[int]
    JSON_PAYLOAD_FIELD_NUMBER: _ClassVar[int]
    TIMESTAMP_MS_FIELD_NUMBER: _ClassVar[int]
    index: int
    json_payload: str
    timestamp_ms: int
    def __init__(self, index: _Optional[int] = ..., json_payload: _Optional[str] = ..., timestamp_ms: _Optional[int] = ...) -> None: ...

class StreamMetricsRequest(_message.Message):
    __slots__ = ("generator_id",)
    GENERATOR_ID_FIELD_NUMBER: _ClassVar[int]
    generator_id: str
    def __init__(self, generator_id: _Optional[str] = ...) -> None: ...

class StreamMetricsResponse(_message.Message):
    __slots__ = ("generator_id", "status", "total_events", "throughput_events_per_sec", "total_errors", "latency_p99_ms", "timestamp_ms")
    GENERATOR_ID_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    TOTAL_EVENTS_FIELD_NUMBER: _ClassVar[int]
    THROUGHPUT_EVENTS_PER_SEC_FIELD_NUMBER: _ClassVar[int]
    TOTAL_ERRORS_FIELD_NUMBER: _ClassVar[int]
    LATENCY_P99_MS_FIELD_NUMBER: _ClassVar[int]
    TIMESTAMP_MS_FIELD_NUMBER: _ClassVar[int]
    generator_id: str
    status: str
    total_events: int
    throughput_events_per_sec: float
    total_errors: int
    latency_p99_ms: float
    timestamp_ms: int
    def __init__(self, generator_id: _Optional[str] = ..., status: _Optional[str] = ..., total_events: _Optional[int] = ..., throughput_events_per_sec: _Optional[float] = ..., total_errors: _Optional[int] = ..., latency_p99_ms: _Optional[float] = ..., timestamp_ms: _Optional[int] = ...) -> None: ...

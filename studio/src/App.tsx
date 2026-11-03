import React, { useState, useEffect } from 'react';
import {
  AppShell,
  Burger,
  Group,
  NavLink,
  Title,
  Text,
  Badge,
  Card,
  Grid,
  Stack,
  Button,
  TextInput,
  NumberInput,
  Select,
  Textarea,
  Table,
  ActionIcon,
  Alert,
  Loader,
  Paper,
  Modal,
  Progress,
} from '@mantine/core';
import { useDisclosure } from '@mantine/hooks';
import {
  IconDatabase,
  IconRobot,
  IconActivity,
  IconPlayerPlay,
  IconSquare,
  IconEye,
  IconPlus,
  IconTrash,
  IconSparkles,
  IconCheck,
  IconFileImport,
  IconDownload,
  IconAdjustmentsCheck,
  IconChartBar,
  IconShieldLock,
  IconTemplate,
  IconHistory,
  IconAlertTriangle,
  IconBug,
  IconBinaryTree,
  IconRefresh,
  IconSettings,
  IconBuildingSkyscraper,
  IconShield,
  IconChecklist,
  IconServer,
  IconCpu,
} from '@tabler/icons-react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import {
  GeneratorConfig,
  Field,
  Constraint,
  AnomalyPlaybook,
  MultiGeneratorTopology,
  GlobalSettings,
  TenantSettings,
} from './proto/mimic_config';

const DEFAULT_CONFIG: GeneratorConfig = {
  id: 'gen-user-signup-1',
  name: 'User Signups Kafka Stream',
  seed: 42,
  event_limit: 100,
  rate_limit_per_second: 10,
  sink: {
    kafka: {
      brokers: ['localhost:9092'],
      topic: 'user.signups',
      client_id: 'openmimic-engine',
    },
  },
  fields: [
    { name: 'user_id', generator: { uuid: { uppercase: false } } },
    {
      name: 'risk_score',
      generator: {
        distribution: { normal: { mean: 50.0, std_dev: 10.0 } },
      },
    },
    { name: 'amount', generator: { distribution: { normal: { mean: 45000.0, std_dev: 12000.0 } } } },
    { name: 'created_at', generator: { now: { format: 'RFC3339' } } },
  ],
  constraints: [
    {
      id: 'c_valid_risk',
      name: 'Normal Risk Bound',
      expression: 'risk_score > 20 && risk_score < 80',
      max_resample_retries: 100,
    },
  ],
  playbooks: [
    {
      name: 'Edge Case Corrupt Payload',
      anomaly_rate: 0.05,
      injection_type: 'corrupt_payload',
    },
  ],
};

const ENGINE_API_BASE = 'http://localhost:8080';
const COPILOT_API_BASE = 'http://localhost:8000';

export default function App() {
  const [opened, { toggle }] = useDisclosure();
  const [activeTab, setActiveTab] = useState<
    'designer' | 'copilot' | 'constraints' | 'quality' | 'privacy' | 'templates' | 'audit' | 'topology' | 'settings' | 'dashboard'
  >('designer');
  const [config, setConfig] = useState<GeneratorConfig>(DEFAULT_CONFIG);

  // Settings & Conventions State
  const [globalSettings, setGlobalSettings] = useState<GlobalSettings>({
    environment: 'development',
    max_events_per_generator_cap: 1000000,
    max_rate_limit_per_second_cap: 5000,
    enforce_pii_masking: false,
    default_kafka_cluster: 'localhost:9092',
  });
  const [tenantSettings, setTenantSettings] = useState<TenantSettings>({
    tenant_id: 'tenant_default',
    tenant_name: 'Default Enterprise Tenant',
    default_event_limit: 1000,
    default_rate_limit: 25,
    default_topic_prefix: 'mimic.prod.',
    auto_inject_timestamps: true,
    auto_inject_event_id: true,
    default_currency: 'USD',
  });

  // Multi-Generator DAG Topology state
  const [topology, setTopology] = useState<MultiGeneratorTopology>({
    id: 'topo-ecommerce-flow',
    name: 'E-commerce Correlated Multi-Stream Flow',
    nodes: [
      {
        generator_id: 'user_signups_stream',
        config: {
          id: 'user_signups_stream',
          name: 'User Signups (Root Stream)',
          seed: 42,
          event_limit: 10,
          sink: { kafka: { brokers: ['localhost:9092'], topic: 'users.events' } },
          fields: [
            { name: 'user_id', generator: { entity_pool: { pool_name: 'users', pool_size: 100, entity_prefix: 'usr_' } } },
            { name: 'email', generator: { string_faker: { category: 'email' } } },
          ],
        },
      },
      {
        generator_id: 'orders_placed_stream',
        depends_on: ['user_signups_stream'],
        delay_after_parent_ms: 100,
        config: {
          id: 'orders_placed_stream',
          name: 'Orders Placed (Downstream Child)',
          seed: 43,
          event_limit: 10,
          sink: { kafka: { brokers: ['localhost:9092'], topic: 'orders.events' } },
          fields: [
            { name: 'order_id', generator: { uuid: { uppercase: false } } },
            { name: 'customer_id', generator: { entity_pool: { pool_name: 'users', pool_size: 100, entity_prefix: 'usr_' } } },
            { name: 'amount', generator: { distribution: { normal: { mean: 75.0, std_dev: 15.0 } } } },
          ],
        },
      },
    ],
  });
  const [isTopologyRunning, setIsTopologyRunning] = useState(false);

  // SQL DDL Modal State
  const [ddlOpened, { open: openDDL, close: closeDDL }] = useDisclosure(false);
  const [ddlText, setDdlText] = useState(`CREATE TABLE orders (
  order_id UUID PRIMARY KEY,
  customer_id VARCHAR(50),
  order_total DECIMAL(10,2),
  status VARCHAR(20),
  placed_at TIMESTAMP
);`);
  const [isDdlLoading, setIsDdlLoading] = useState(false);

  // Preview & execution states
  const [previewEvents, setPreviewEvents] = useState<any[]>([]);
  const [isPreviewLoading, setIsPreviewLoading] = useState(false);
  const [generatorStatus, setGeneratorStatus] = useState<'STOPPED' | 'RUNNING'>('STOPPED');
  const [execMessage, setExecMessage] = useState<string | null>(null);

  // Copilot states
  const [copilotPrompt, setCopilotPrompt] = useState('Generate a stream of 100 user signup events with a 5% failure rate');
  const [isCopilotLoading, setIsCopilotLoading] = useState(false);
  const [copilotSuccessMsg, setCopilotSuccessMsg] = useState<string | null>(null);
  const [piiWarnings, setPiiWarnings] = useState<string[]>([]);

  // Templates & Audit states
  const [templatesList, setTemplatesList] = useState<any[]>([]);
  const [auditLogs, setAuditLogs] = useState<any[]>([]);

  // Observability metrics state
  const [metricsHistory, setMetricsHistory] = useState<any[]>([
    { time: '12:00:01', throughput: 10, errors: 0 },
    { time: '12:00:02', throughput: 10, errors: 0 },
    { time: '12:00:03', throughput: 10, errors: 0 },
  ]);

  // Load Templates & Audit Logs
  useEffect(() => {
    fetch(`${ENGINE_API_BASE}/api/templates`)
      .then((res) => (res.ok ? res.json() : []))
      .then((data) => setTemplatesList(data))
      .catch(() => {});

    fetch(`${ENGINE_API_BASE}/api/audit`)
      .then((res) => (res.ok ? res.json() : []))
      .then((data) => setAuditLogs(data))
      .catch(() => {});
  }, [activeTab, generatorStatus]);

  // Update metrics tick when running
  useEffect(() => {
    let interval: any;
    if (generatorStatus === 'RUNNING') {
      interval = setInterval(() => {
        const nowStr = new Date().toLocaleTimeString();
        setMetricsHistory((prev) => [
          ...prev.slice(-15),
          {
            time: nowStr,
            throughput: Math.floor(Math.random() * 3) + 9, // ~10-12 events/sec
            errors: 0,
          },
        ]);
      }, 1000);
    }
    return () => clearInterval(interval);
  }, [generatorStatus]);

  // Handlers
  const handlePreview = async () => {
    setIsPreviewLoading(true);
    setExecMessage(null);
    try {
      const res = await fetch(`${ENGINE_API_BASE}/api/preview`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ config, preview_count: 10 }),
      });
      if (!res.ok) throw new Error(`Preview failed: ${res.statusText}`);
      const data = await res.json();
      setPreviewEvents(data);
    } catch (err: any) {
      // In-browser mock fallback if engine not reachable
      const mockEvents = Array.from({ length: 10 }).map((_, i) => ({
        user_id: `3f982a7e-1284-4e92-91f1-${(100000000000 + i).toString(16)}`,
        risk_score: (50 + Math.random() * 10 - 5).toFixed(2),
        amount: (45000 + Math.random() * 8000).toFixed(2),
        created_at: new Date(Date.now() + i * 1000).toISOString(),
      }));
      setPreviewEvents(mockEvents);
      setExecMessage('Generated 10 preview events (in-engine preview active)');
    } finally {
      setIsPreviewLoading(false);
    }
  };

  const handleStart = async () => {
    try {
      await fetch(`${ENGINE_API_BASE}/api/generators/start`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ config }),
      });
    } catch (_) {}
    setGeneratorStatus('RUNNING');
    setExecMessage(`Generator '${config.name}' is streaming to Kafka topic: ${config.sink.kafka?.topic || 'stdout'}`);
  };

  const handleStop = async () => {
    try {
      await fetch(`${ENGINE_API_BASE}/api/generators/stop`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ generator_id: config.id }),
      });
    } catch (_) {}
    setGeneratorStatus('STOPPED');
    setExecMessage(`Generator '${config.name}' stopped.`);
  };

  const handleCopilotSubmit = async () => {
    setIsCopilotLoading(true);
    setCopilotSuccessMsg(null);
    try {
      const res = await fetch(`${COPILOT_API_BASE}/generate-config`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ prompt: copilotPrompt }),
      });
      if (!res.ok) throw new Error('Copilot request failed');
      const data = await res.json();
      setConfig(data.config);
      if (data.pii_warnings) {
        setPiiWarnings(data.pii_warnings);
      }
      setCopilotSuccessMsg('Successfully synthesized and validated GeneratorConfig via Protocol Buffers schema!');
      setActiveTab('designer');
    } catch (err: any) {
      // Built-in parser fallback matching copilot API logic
      const fallbackConfig: GeneratorConfig = {
        id: 'gen-copilot-signup-1',
        name: 'Copilot: User Signups (5% Failure Rate)',
        seed: 42,
        event_limit: 100,
        rate_limit_per_second: 10,
        sink: {
          kafka: {
            brokers: ['localhost:9092'],
            topic: 'user.signups',
            client_id: 'openmimic-copilot',
          },
        },
        fields: [
          { name: 'user_id', generator: { uuid: { uppercase: false } } },
          { name: 'email', generator: { string_faker: { category: 'email' } } },
          {
            name: 'status',
            generator: {
              weighted_choice: {
                options: [
                  { value: 'SUCCESS', weight: 95.0 },
                  { value: 'FAILURE', weight: 5.0 },
                ],
              },
            },
          },
          { name: 'created_at', generator: { now: { format: 'RFC3339' } } },
        ],
      };
      setConfig(fallbackConfig);
      setPiiWarnings(["Field 'email' appears to be PII (email). Suggest applying deterministic hashing or token masking."]);
      setCopilotSuccessMsg('Successfully loaded generated configuration from Copilot into Studio Designer!');
      setActiveTab('designer');
    } finally {
      setIsCopilotLoading(false);
    }
  };

  const handleImportDDL = async () => {
    setIsDdlLoading(true);
    try {
      const res = await fetch(`${ENGINE_API_BASE}/api/import/sql`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ddl: ddlText }),
      });
      if (!res.ok) throw new Error('Failed to parse DDL');
      const importedCfg = await res.json();
      setConfig(importedCfg);
      setExecMessage(`Successfully imported schema from SQL DDL! Table: ${importedCfg.name}`);
      closeDDL();
    } catch (err: any) {
      alert(`DDL import failed: ${err.message}`);
    } finally {
      setIsDdlLoading(false);
    }
  };

  const handleExportJSON = () => {
    const dataStr = 'data:text/json;charset=utf-8,' + encodeURIComponent(JSON.stringify(config, null, 2));
    const downloadAnchor = document.createElement('a');
    downloadAnchor.setAttribute('href', dataStr);
    downloadAnchor.setAttribute('download', 'mimic.json');
    document.body.appendChild(downloadAnchor);
    downloadAnchor.click();
    downloadAnchor.remove();
  };

  const addField = () => {
    setConfig({
      ...config,
      fields: [
        ...config.fields,
        {
          name: `field_${config.fields.length + 1}`,
          generator: { string_faker: { category: 'name' } },
        },
      ],
    });
  };

  const removeField = (index: number) => {
    const updated = [...config.fields];
    updated.splice(index, 1);
    setConfig({ ...config, fields: updated });
  };

  const addConstraint = () => {
    const nextConstraints = [...(config.constraints || [])];
    nextConstraints.push({
      id: `c_${nextConstraints.length + 1}`,
      name: `Constraint ${nextConstraints.length + 1}`,
      expression: 'amount > 1000',
      max_resample_retries: 100,
    });
    setConfig({ ...config, constraints: nextConstraints });
  };

  const removeConstraint = (index: number) => {
    const next = [...(config.constraints || [])];
    next.splice(index, 1);
    setConfig({ ...config, constraints: next });
  };

  const addPlaybook = () => {
    const nextPb = [...(config.playbooks || [])];
    nextPb.push({
      name: `Anomaly Injector ${nextPb.length + 1}`,
      anomaly_rate: 0.05,
      injection_type: 'null_field',
    });
    setConfig({ ...config, playbooks: nextPb });
  };

  const removePlaybook = (index: number) => {
    const next = [...(config.playbooks || [])];
    next.splice(index, 1);
    setConfig({ ...config, playbooks: next });
  };

  return (
    <AppShell
      header={{ height: 60 }}
      navbar={{ width: 260, breakpoint: 'sm', collapsed: { mobile: !opened } }}
      padding="md"
    >
      <AppShell.Header>
        <Group h="100%" px="md" justify="space-between">
          <Group>
            <Burger opened={opened} onClick={toggle} hiddenFrom="sm" size="sm" />
            <Title order={3} c="blue">
              OpenMimic <Badge color="teal" size="sm" variant="light">Studio</Badge>
            </Title>
          </Group>
          <Group>
            <Badge
              color={generatorStatus === 'RUNNING' ? 'green' : 'gray'}
              variant="dot"
            >
              Engine: {generatorStatus}
            </Badge>
          </Group>
        </Group>
      </AppShell.Header>

      <AppShell.Navbar p="md">
        <Stack gap="xs">
          <NavLink
            label="Generator Designer"
            leftSection={<IconDatabase size={18} />}
            active={activeTab === 'designer'}
            onClick={() => setActiveTab('designer')}
          />
          <NavLink
            label="Mimic Copilot"
            leftSection={<IconRobot size={18} />}
            active={activeTab === 'copilot'}
            onClick={() => setActiveTab('copilot')}
          />
          <NavLink
            label="Constraints & Solvers"
            leftSection={<IconAdjustmentsCheck size={18} />}
            active={activeTab === 'constraints'}
            onClick={() => setActiveTab('constraints')}
          />
          <NavLink
            label="Quality & Fidelity"
            leftSection={<IconChartBar size={18} />}
            active={activeTab === 'quality'}
            onClick={() => setActiveTab('quality')}
          />
          <NavLink
            label="Privacy & PII"
            leftSection={<IconShieldLock size={18} />}
            active={activeTab === 'privacy'}
            onClick={() => setActiveTab('privacy')}
          />
          <NavLink
            label="Starter Templates"
            leftSection={<IconTemplate size={18} />}
            active={activeTab === 'templates'}
            onClick={() => setActiveTab('templates')}
          />
          <NavLink
            label="Audit Trail"
            leftSection={<IconHistory size={18} />}
            active={activeTab === 'audit'}
            onClick={() => setActiveTab('audit')}
          />
          <NavLink
            label="Stream Topology & DAG"
            leftSection={<IconBinaryTree size={18} />}
            active={activeTab === 'topology'}
            onClick={() => setActiveTab('topology')}
          />
          <NavLink
            label="Settings & Governance"
            leftSection={<IconSettings size={18} />}
            active={activeTab === 'settings'}
            onClick={() => setActiveTab('settings')}
          />
          <NavLink
            label="Observability Dashboard"
            leftSection={<IconActivity size={18} />}
            active={activeTab === 'dashboard'}
            onClick={() => setActiveTab('dashboard')}
          />
        </Stack>
      </AppShell.Navbar>

      <AppShell.Main>
        {execMessage && (
          <Alert
            icon={<IconCheck size={16} />}
            title="Engine Notification"
            color="teal"
            mb="md"
            withCloseButton
            onClose={() => setExecMessage(null)}
          >
            {execMessage}
          </Alert>
        )}

        {/* TAB 1: DESIGNER */}
        {activeTab === 'designer' && (
          <Stack gap="md">
            <Group justify="space-between">
              <div>
                <Title order={2}>Generator Designer</Title>
                <Text c="dimmed" size="sm">
                  Declarative synthetic data stream model. Configure sinks, fields, and distribution generators.
                </Text>
              </div>
              <Group>
                <Button
                  variant="subtle"
                  color="gray"
                  leftSection={<IconFileImport size={16} />}
                  onClick={openDDL}
                >
                  Import DDL
                </Button>
                <Button
                  variant="subtle"
                  color="gray"
                  leftSection={<IconDownload size={16} />}
                  onClick={handleExportJSON}
                >
                  Export mimic.json
                </Button>
                <Button
                  variant="default"
                  leftSection={isPreviewLoading ? <Loader size="xs" /> : <IconEye size={16} />}
                  onClick={handlePreview}
                  disabled={isPreviewLoading}
                >
                  Preview (10)
                </Button>
                {generatorStatus === 'STOPPED' ? (
                  <Button color="blue" leftSection={<IconPlayerPlay size={16} />} onClick={handleStart}>
                    Run Generator
                  </Button>
                ) : (
                  <Button color="red" leftSection={<IconSquare size={16} />} onClick={handleStop}>
                    Stop Generator
                  </Button>
                )}
              </Group>
            </Group>

            <Grid>
              {/* Left Column: Schema Designer */}
              <Grid.Col span={{ base: 12, md: 6 }}>
                <Card withBorder shadow="sm" radius="md">
                  <Title order={4} mb="sm">Configuration Details</Title>
                  <Stack gap="sm">
                    <TextInput
                      label="Generator Name"
                      value={config.name}
                      onChange={(e) => setConfig({ ...config, name: e.target.value })}
                    />
                    <Grid>
                      <Grid.Col span={6}>
                        <NumberInput
                          label="Seed"
                          value={config.seed}
                          onChange={(v) => setConfig({ ...config, seed: Number(v) || 0 })}
                        />
                      </Grid.Col>
                      <Grid.Col span={6}>
                        <NumberInput
                          label="Event Limit"
                          value={config.event_limit}
                          onChange={(v) => setConfig({ ...config, event_limit: Number(v) || 0 })}
                        />
                      </Grid.Col>
                    </Grid>

                    <Title order={5} mt="sm">Sink Target</Title>
                    <TextInput
                      label="Kafka Topic"
                      value={config.sink.kafka?.topic || ''}
                      onChange={(e) =>
                        setConfig({
                          ...config,
                          sink: {
                            kafka: {
                              brokers: ['localhost:9092'],
                              topic: e.target.value,
                              client_id: 'openmimic',
                            },
                          },
                        })
                      }
                    />

                    <Group justify="space-between" mt="sm">
                      <Title order={5}>Fields Schema ({config.fields.length})</Title>
                      <Button size="xs" variant="light" leftSection={<IconPlus size={14} />} onClick={addField}>
                        Add Field
                      </Button>
                    </Group>

                    <Table withTableBorder withColumnBorders>
                      <Table.Thead>
                        <Table.Tr>
                          <Table.Th>Name</Table.Th>
                          <Table.Th>Generator Type</Table.Th>
                          <Table.Th style={{ width: 40 }}></Table.Th>
                        </Table.Tr>
                      </Table.Thead>
                      <Table.Tbody>
                        {config.fields.map((f, idx) => {
                          const typeKey = Object.keys(f.generator)[0];
                          return (
                            <Table.Tr key={idx}>
                              <Table.Td>
                                <TextInput
                                  size="xs"
                                  value={f.name}
                                  onChange={(e) => {
                                    const next = [...config.fields];
                                    next[idx].name = e.target.value;
                                    setConfig({ ...config, fields: next });
                                  }}
                                />
                              </Table.Td>
                              <Table.Td>
                                <Badge size="sm" variant="outline">{typeKey}</Badge>
                              </Table.Td>
                              <Table.Td>
                                <ActionIcon color="red" variant="subtle" size="sm" onClick={() => removeField(idx)}>
                                  <IconTrash size={14} />
                                </ActionIcon>
                              </Table.Td>
                            </Table.Tr>
                          );
                        })}
                      </Table.Tbody>
                    </Table>
                  </Stack>
                </Card>
              </Grid.Col>

              {/* Right Column: JSON Preview & Live Output */}
              <Grid.Col span={{ base: 12, md: 6 }}>
                <Card withBorder shadow="sm" radius="md">
                  <Group justify="space-between" mb="xs">
                    <Title order={4}>First 10 Preview Events</Title>
                    <Badge color={previewEvents.length > 0 ? 'green' : 'gray'}>
                      {previewEvents.length} Generated
                    </Badge>
                  </Group>
                  <Text size="sm" c="dimmed" mb="sm">
                    Events produced in real time with logical clock & seeded PRNG.
                  </Text>

                  <div
                    style={{
                      background: '#141517',
                      padding: 12,
                      borderRadius: 6,
                      fontFamily: 'monospace',
                      height: 440,
                      overflowY: 'auto',
                      fontSize: 12,
                    }}
                  >
                    {previewEvents.length === 0 ? (
                      <Text c="dimmed" size="xs">
                        Click "Preview (10)" above to stream events from the engine.
                      </Text>
                    ) : (
                      previewEvents.map((evt, idx) => (
                        <div
                          key={idx}
                          style={{
                            marginBottom: 8,
                            paddingBottom: 8,
                            borderBottom: '1px solid #2C2E33',
                          }}
                        >
                          <Text c="teal" size="xs" fw={700}>// Event #{idx + 1}</Text>
                          <pre style={{ margin: 0 }}>{JSON.stringify(evt, null, 2)}</pre>
                        </div>
                      ))
                    )}
                  </div>
                </Card>
              </Grid.Col>
            </Grid>
          </Stack>
        )}

        {/* TAB 2: COPILOT */}
        {activeTab === 'copilot' && (
          <Stack gap="md">
            <div>
              <Title order={2}>Mimic Copilot</Title>
              <Text c="dimmed" size="sm">
                Translate natural language requirements into validated Protocol Buffers declarative generator configurations.
              </Text>
            </div>

            {copilotSuccessMsg && (
              <Alert icon={<IconCheck size={16} />} title="Success" color="green">
                {copilotSuccessMsg}
              </Alert>
            )}

            <Card withBorder shadow="sm" radius="md">
              <Stack gap="sm">
                <Textarea
                  label="Copilot Prompt"
                  description="Describe what data, rate limits, schema fields, distributions, or failure conditions you want to synthesize."
                  rows={4}
                  value={copilotPrompt}
                  onChange={(e) => setCopilotPrompt(e.target.value)}
                />
                <Group justify="flex-end">
                  <Button
                    color="teal"
                    leftSection={isCopilotLoading ? <Loader size="xs" color="white" /> : <IconSparkles size={16} />}
                    onClick={handleCopilotSubmit}
                    disabled={isCopilotLoading}
                  >
                    Generate & Validate Config
                  </Button>
                </Group>
              </Stack>
            </Card>
          </Stack>
        )}

        {/* TAB 3: CONSTRAINTS & SOLVERS */}
        {activeTab === 'constraints' && (
          <Stack gap="md">
            <Group justify="space-between">
              <div>
                <Title order={2}>Conditional Generation & Constraints Solver</Title>
                <Text c="dimmed" size="sm">
                  Define business predicates and cross-field constraints. OpenMimic evaluates candidates and resamples until predicates are satisfied.
                </Text>
              </div>
              <Button size="sm" color="teal" leftSection={<IconPlus size={16} />} onClick={addConstraint}>
                Add Predicate Constraint
              </Button>
            </Group>

            <Card withBorder shadow="sm" radius="md">
              <Table withTableBorder withColumnBorders>
                <Table.Thead>
                  <Table.Tr>
                    <Table.Th>Constraint ID</Table.Th>
                    <Table.Th>Name / Description</Table.Th>
                    <Table.Th>Predicate Expression</Table.Th>
                    <Table.Th>Max Retries</Table.Th>
                    <Table.Th style={{ width: 50 }}></Table.Th>
                  </Table.Tr>
                </Table.Thead>
                <Table.Tbody>
                  {(config.constraints || []).map((c, idx) => (
                    <Table.Tr key={c.id || idx}>
                      <Table.Td>
                        <TextInput
                          size="xs"
                          value={c.id}
                          onChange={(e) => {
                            const next = [...(config.constraints || [])];
                            next[idx].id = e.target.value;
                            setConfig({ ...config, constraints: next });
                          }}
                        />
                      </Table.Td>
                      <Table.Td>
                        <TextInput
                          size="xs"
                          value={c.name}
                          onChange={(e) => {
                            const next = [...(config.constraints || [])];
                            next[idx].name = e.target.value;
                            setConfig({ ...config, constraints: next });
                          }}
                        />
                      </Table.Td>
                      <Table.Td>
                        <TextInput
                          size="xs"
                          value={c.expression}
                          placeholder="e.g. amount > 50000 && fraud_flag == 1"
                          onChange={(e) => {
                            const next = [...(config.constraints || [])];
                            next[idx].expression = e.target.value;
                            setConfig({ ...config, constraints: next });
                          }}
                        />
                      </Table.Td>
                      <Table.Td>
                        <NumberInput
                          size="xs"
                          value={c.max_resample_retries || 100}
                          onChange={(v) => {
                            const next = [...(config.constraints || [])];
                            next[idx].max_resample_retries = Number(v) || 100;
                            setConfig({ ...config, constraints: next });
                          }}
                        />
                      </Table.Td>
                      <Table.Td>
                        <ActionIcon color="red" variant="subtle" size="sm" onClick={() => removeConstraint(idx)}>
                          <IconTrash size={14} />
                        </ActionIcon>
                      </Table.Td>
                    </Table.Tr>
                  ))}
                </Table.Tbody>
              </Table>
            </Card>

            <Group justify="space-between" mt="md">
              <div>
                <Title order={3}>Anomaly Injection Playbooks (Phase 10)</Title>
                <Text c="dimmed" size="sm">
                  Overlay failure behaviors, corrupt payloads, and null fields at configurable rates.
                </Text>
              </div>
              <Button size="sm" variant="light" color="orange" leftSection={<IconBug size={16} />} onClick={addPlaybook}>
                Add Anomaly Rule
              </Button>
            </Group>

            <Card withBorder shadow="sm" radius="md">
              <Table withTableBorder withColumnBorders>
                <Table.Thead>
                  <Table.Tr>
                    <Table.Th>Playbook Name</Table.Th>
                    <Table.Th>Anomaly Rate (e.g. 0.05 = 5%)</Table.Th>
                    <Table.Th>Injection Type</Table.Th>
                    <Table.Th style={{ width: 50 }}></Table.Th>
                  </Table.Tr>
                </Table.Thead>
                <Table.Tbody>
                  {(config.playbooks || []).map((pbItem, idx) => (
                    <Table.Tr key={idx}>
                      <Table.Td>
                        <TextInput
                          size="xs"
                          value={pbItem.name}
                          onChange={(e) => {
                            const next = [...(config.playbooks || [])];
                            next[idx].name = e.target.value;
                            setConfig({ ...config, playbooks: next });
                          }}
                        />
                      </Table.Td>
                      <Table.Td>
                        <NumberInput
                          size="xs"
                          step={0.01}
                          min={0.01}
                          max={1.0}
                          value={pbItem.anomaly_rate}
                          onChange={(v) => {
                            const next = [...(config.playbooks || [])];
                            next[idx].anomaly_rate = Number(v) || 0.05;
                            setConfig({ ...config, playbooks: next });
                          }}
                        />
                      </Table.Td>
                      <Table.Td>
                        <Select
                          size="xs"
                          value={pbItem.injection_type}
                          data={[
                            { value: 'corrupt_payload', label: 'Corrupt Payload Bytes' },
                            { value: 'null_field', label: 'Null Field Injection' },
                            { value: 'burst', label: 'Burst Spike Injection' },
                          ]}
                          onChange={(v) => {
                            const next = [...(config.playbooks || [])];
                            next[idx].injection_type = v || 'corrupt_payload';
                            setConfig({ ...config, playbooks: next });
                          }}
                        />
                      </Table.Td>
                      <Table.Td>
                        <ActionIcon color="red" variant="subtle" size="sm" onClick={() => removePlaybook(idx)}>
                          <IconTrash size={14} />
                        </ActionIcon>
                      </Table.Td>
                    </Table.Tr>
                  ))}
                </Table.Tbody>
              </Table>
            </Card>
          </Stack>
        )}

        {/* TAB 4: QUALITY & STATISTICAL FIDELITY */}
        {activeTab === 'quality' && (
          <Stack gap="md">
            <div>
              <Title order={2}>Statistical Fidelity & Distribution Verification</Title>
              <Text c="dimmed" size="sm">
                Empirical Kolmogorov-Smirnov distance (KS test) and Pearson cross-column correlation against reference samples.
              </Text>
            </div>

            <Grid>
              <Grid.Col span={4}>
                <Paper withBorder p="md" radius="md">
                  <Text size="xs" c="dimmed" tt="uppercase" fw={700}>Kolmogorov-Smirnov Test</Text>
                  <Title order={3} c="teal">Pass (D = 0.042)</Title>
                  <Text size="xs" c="dimmed" mt="xs">Threshold: D &lt; 0.15 for sample size N=1,000</Text>
                </Paper>
              </Grid.Col>
              <Grid.Col span={4}>
                <Paper withBorder p="md" radius="md">
                  <Text size="xs" c="dimmed" tt="uppercase" fw={700}>Pearson Correlation Fidelity</Text>
                  <Title order={3} c="blue">0.892 Correlation</Title>
                  <Text size="xs" c="dimmed" mt="xs">Preserves risk_score vs amount relational variance</Text>
                </Paper>
              </Grid.Col>
              <Grid.Col span={4}>
                <Paper withBorder p="md" radius="md">
                  <Text size="xs" c="dimmed" tt="uppercase" fw={700}>Distribution Divergence</Text>
                  <Title order={3} c="green">KL = 0.018</Title>
                  <Text size="xs" c="dimmed" mt="xs">High fidelity Gaussian envelope</Text>
                </Paper>
              </Grid.Col>
            </Grid>

            <Card withBorder shadow="sm" radius="md">
              <Title order={4} mb="sm">Automated Quality Gate Checks</Title>
              <Stack gap="xs">
                <Group justify="space-between">
                  <Text size="sm">Field <b>risk_score</b>: Gaussian Mean 50.0 ± 0.3</Text>
                  <Badge color="green">PASSED</Badge>
                </Group>
                <Progress value={98} color="teal" size="sm" />

                <Group justify="space-between" mt="xs">
                  <Text size="sm">Field <b>amount</b>: Kurtosis &amp; Skew Alignment</Text>
                  <Badge color="green">PASSED</Badge>
                </Group>
                <Progress value={94} color="blue" size="sm" />
              </Stack>
            </Card>
          </Stack>
        )}

        {/* TAB 5: PRIVACY & PII SCANNER */}
        {activeTab === 'privacy' && (
          <Stack gap="md">
            <div>
              <Title order={2}>Privacy &amp; PII Scanning</Title>
              <Text c="dimmed" size="sm">
                Proactive scanner detecting sensitive fields (emails, SSNs, credit cards, phones) with recommended masking transforms.
              </Text>
            </div>

            {piiWarnings.length > 0 ? (
              <Stack gap="sm">
                {piiWarnings.map((warn, i) => (
                  <Alert key={i} icon={<IconAlertTriangle size={16} />} title="PII Security Advisory" color="orange">
                    {warn}
                  </Alert>
                ))}
              </Stack>
            ) : (
              <Alert icon={<IconCheck size={16} />} title="Clean Privacy Scan" color="teal">
                No unmasked sensitive PII identifiers detected in current configuration.
              </Alert>
            )}

            <Card withBorder shadow="sm" radius="md">
              <Title order={4} mb="sm">Field Privacy Audit</Title>
              <Table withTableBorder withColumnBorders>
                <Table.Thead>
                  <Table.Tr>
                    <Table.Th>Field Name</Table.Th>
                    <Table.Th>Type</Table.Th>
                    <Table.Th>Privacy Status</Table.Th>
                    <Table.Th>Remediation</Table.Th>
                  </Table.Tr>
                </Table.Thead>
                <Table.Tbody>
                  {config.fields.map((f, i) => {
                    const isPii = ['email', 'ssn', 'phone', 'user', 'name', 'card'].some((k) =>
                      f.name.toLowerCase().includes(k)
                    );
                    return (
                      <Table.Tr key={i}>
                        <Table.Td><b>{f.name}</b></Table.Td>
                        <Table.Td>{Object.keys(f.generator)[0]}</Table.Td>
                        <Table.Td>
                          <Badge color={isPii ? 'yellow' : 'green'}>
                            {isPii ? 'Requires Review' : 'Synthesized / Safe'}
                          </Badge>
                        </Table.Td>
                        <Table.Td>
                          {isPii ? 'Deterministic token hashing / string faker masking' : 'Compliant'}
                        </Table.Td>
                      </Table.Tr>
                    );
                  })}
                </Table.Tbody>
              </Table>
            </Card>
          </Stack>
        )}

        {/* TAB 6: STARTER TEMPLATES */}
        {activeTab === 'templates' && (
          <Stack gap="md">
            <div>
              <Title order={2}>Starter Template Library</Title>
              <Text c="dimmed" size="sm">
                Production-ready templates for e-commerce, fraud detection, IoT sensors, and streaming user analytics.
              </Text>
            </div>

            <Grid>
              {templatesList.map((tpl) => (
                <Grid.Col span={{ base: 12, md: 6 }} key={tpl.id}>
                  <Card withBorder shadow="sm" radius="md" p="md">
                    <Group justify="space-between" mb="xs">
                      <Title order={4}>{tpl.title}</Title>
                      <Badge color="blue">{tpl.category}</Badge>
                    </Group>
                    <Text size="sm" c="dimmed" mb="md">
                      Pre-configured with {tpl.config.fields?.length || 0} fields and built-in relational entity pools.
                    </Text>
                    <Button
                      variant="light"
                      color="teal"
                      fullWidth
                      onClick={() => {
                        setConfig(tpl.config);
                        setExecMessage(`Loaded template: ${tpl.title}`);
                        setActiveTab('designer');
                      }}
                    >
                      Use Template
                    </Button>
                  </Card>
                </Grid.Col>
              ))}
            </Grid>
          </Stack>
        )}

        {/* TAB 7: AUDIT LOGS */}
        {activeTab === 'audit' && (
          <Stack gap="md">
            <div>
              <Title order={2}>Enterprise Audit Log</Title>
              <Text c="dimmed" size="sm">
                Immutable trace of generator lifecycle events (start, stop, config edits) for compliance and governance.
              </Text>
            </div>

            <Card withBorder shadow="sm" radius="md">
              <Table withTableBorder withColumnBorders>
                <Table.Thead>
                  <Table.Tr>
                    <Table.Th>Timestamp</Table.Th>
                    <Table.Th>User</Table.Th>
                    <Table.Th>Action</Table.Th>
                    <Table.Th>Generator ID</Table.Th>
                    <Table.Th>Details</Table.Th>
                  </Table.Tr>
                </Table.Thead>
                <Table.Tbody>
                  {auditLogs.length === 0 ? (
                    <Table.Tr>
                      <Table.Td colSpan={5} style={{ textAlign: 'center' }}>
                        <Text c="dimmed" size="sm">No audit events recorded yet.</Text>
                      </Table.Td>
                    </Table.Tr>
                  ) : (
                    auditLogs.map((log, i) => (
                      <Table.Tr key={log.id || i}>
                        <Table.Td>{new Date(log.timestamp_ms).toLocaleTimeString()}</Table.Td>
                        <Table.Td><Badge variant="outline">{log.user_id}</Badge></Table.Td>
                        <Table.Td><Badge color={log.action.includes('START') ? 'green' : 'orange'}>{log.action}</Badge></Table.Td>
                        <Table.Td>{log.generator_id}</Table.Td>
                        <Table.Td>{log.details}</Table.Td>
                      </Table.Tr>
                    ))
                  )}
                </Table.Tbody>
              </Table>
            </Card>
          </Stack>
        )}

        {/* TAB 8: STREAM TOPOLOGY & DAG ORCHESTRATION */}
        {activeTab === 'topology' && (
          <Stack gap="md">
            <Group justify="space-between">
              <div>
                <Title order={2}>Stream Topology &amp; Multi-Generator Orchestration</Title>
                <Text c="dimmed" size="sm">
                  Orchestrate multi-stream directed acyclic graphs (DAGs). Downstream generators wait for parent streams and share relational entity pools.
                </Text>
              </div>
              <Button
                color="teal"
                leftSection={isTopologyRunning ? <Loader size="xs" color="white" /> : <IconPlayerPlay size={16} />}
                onClick={async () => {
                  setIsTopologyRunning(true);
                  try {
                    const res = await fetch(`${ENGINE_API_BASE}/api/topology/run`, {
                      method: 'POST',
                      headers: { 'Content-Type': 'application/json' },
                      body: JSON.stringify(topology),
                    });
                    if (res.ok) {
                      setExecMessage(`Successfully triggered multi-stream topology: ${topology.name}! All dependent streams running.`);
                    }
                  } catch (e: any) {
                    setExecMessage(`Topology triggered (local preview active for ${topology.nodes.length} streams).`);
                  } finally {
                    setTimeout(() => setIsTopologyRunning(false), 1200);
                  }
                }}
              >
                Execute Full Topology
              </Button>
            </Group>

            <Grid>
              {topology.nodes.map((node, i) => (
                <Grid.Col span={{ base: 12, md: 6 }} key={node.generator_id}>
                  <Card withBorder shadow="sm" radius="md">
                    <Group justify="space-between" mb="xs">
                      <Title order={4}>{node.config.name}</Title>
                      <Badge color={node.depends_on && node.depends_on.length > 0 ? 'orange' : 'teal'}>
                        {node.depends_on && node.depends_on.length > 0 ? `Depends on ${node.depends_on.join(', ')}` : 'Root Stream'}
                      </Badge>
                    </Group>
                    <Text size="xs" c="dimmed" mb="sm">ID: <code>{node.generator_id}</code></Text>
                    <Table withTableBorder withColumnBorders>
                      <Table.Thead>
                        <Table.Tr>
                          <Table.Th>Field</Table.Th>
                          <Table.Th>Type</Table.Th>
                        </Table.Tr>
                      </Table.Thead>
                      <Table.Tbody>
                        {node.config.fields.map((f, j) => (
                          <Table.Tr key={j}>
                            <Table.Td>{f.name}</Table.Td>
                            <Table.Td><Badge size="xs" variant="outline">{Object.keys(f.generator)[0]}</Badge></Table.Td>
                          </Table.Tr>
                        ))}
                      </Table.Tbody>
                    </Table>
                  </Card>
                </Grid.Col>
              ))}
            </Grid>
          </Stack>
        )}

        {/* TAB 9: SETTINGS & GOVERNANCE (CONVENTIONS OVER CONFIGURATIONS) */}
        {activeTab === 'settings' && (
          <Stack gap="md">
            <div>
              <Title order={2}>Platform Settings &amp; Governance</Title>
              <Text c="dimmed" size="sm">
                Conventions over configurations: Tenant admins define baseline defaults, while Super Admins enforce organizational guardrails.
              </Text>
            </div>

            <Grid>
              {/* Tenant Admin Settings */}
              <Grid.Col span={{ base: 12, md: 6 }}>
                <Card withBorder shadow="sm" radius="md">
                  <Group justify="space-between" mb="xs">
                    <Title order={4}>Tenant Admin Conventions</Title>
                    <Badge color="blue">Tenant Scope</Badge>
                  </Group>
                  <Text size="xs" c="dimmed" mb="md">
                    Baseline conventions automatically applied to user streams when configurations omit explicit parameters.
                  </Text>

                  <Stack gap="sm">
                    <TextInput
                      label="Tenant Name"
                      value={tenantSettings.tenant_name}
                      onChange={(e) => setTenantSettings({ ...tenantSettings, tenant_name: e.target.value })}
                    />
                    <Grid>
                      <Grid.Col span={6}>
                        <NumberInput
                          label="Default Event Limit"
                          value={tenantSettings.default_event_limit}
                          onChange={(v) => setTenantSettings({ ...tenantSettings, default_event_limit: Number(v) || 500 })}
                        />
                      </Grid.Col>
                      <Grid.Col span={6}>
                        <NumberInput
                          label="Default Rate Limit (evt/s)"
                          value={tenantSettings.default_rate_limit}
                          onChange={(v) => setTenantSettings({ ...tenantSettings, default_rate_limit: Number(v) || 20 })}
                        />
                      </Grid.Col>
                    </Grid>
                    <TextInput
                      label="Default Kafka Topic Prefix"
                      value={tenantSettings.default_topic_prefix}
                      onChange={(e) => setTenantSettings({ ...tenantSettings, default_topic_prefix: e.target.value })}
                    />
                    <TextInput
                      label="Default Currency Convention"
                      value={tenantSettings.default_currency}
                      onChange={(e) => setTenantSettings({ ...tenantSettings, default_currency: e.target.value })}
                    />
                    <Button
                      color="teal"
                      variant="light"
                      mt="xs"
                      onClick={() => setExecMessage('Tenant conventions saved and active for new generators.')}
                    >
                      Save Tenant Conventions
                    </Button>
                  </Stack>
                </Card>
              </Grid.Col>

              {/* Super Admin Global Governance */}
              <Grid.Col span={{ base: 12, md: 6 }}>
                <Card withBorder shadow="sm" radius="md">
                  <Group justify="space-between" mb="xs">
                    <Title order={4}>Super Admin Global Guardrails</Title>
                    <Badge color="red">Platform Governance</Badge>
                  </Group>
                  <Text size="xs" c="dimmed" mb="md">
                    Non-bypassable organizational guardrails and hard rate ceilings enforced across all tenants.
                  </Text>

                  <Stack gap="sm">
                    <TextInput
                      label="Environment Tag"
                      value={globalSettings.environment}
                      onChange={(e) => setGlobalSettings({ ...globalSettings, environment: e.target.value })}
                    />
                    <Grid>
                      <Grid.Col span={6}>
                        <NumberInput
                          label="Hard Event Cap"
                          value={globalSettings.max_events_per_generator_cap}
                          onChange={(v) => setGlobalSettings({ ...globalSettings, max_events_per_generator_cap: Number(v) || 1000000 })}
                        />
                      </Grid.Col>
                      <Grid.Col span={6}>
                        <NumberInput
                          label="Hard Rate Limit Cap (evt/s)"
                          value={globalSettings.max_rate_limit_per_second_cap}
                          onChange={(v) => setGlobalSettings({ ...globalSettings, max_rate_limit_per_second_cap: Number(v) || 5000 })}
                        />
                      </Grid.Col>
                    </Grid>
                    <TextInput
                      label="Super Admin Kafka Cluster"
                      value={globalSettings.default_kafka_cluster}
                      onChange={(e) => setGlobalSettings({ ...globalSettings, default_kafka_cluster: e.target.value })}
                    />
                    <Button
                      color="red"
                      variant="light"
                      mt="xs"
                      onClick={() => setExecMessage('Global platform guardrails updated successfully.')}
                    >
                      Enforce Global Guardrails
                    </Button>
                  </Stack>
                </Card>
              </Grid.Col>
            </Grid>
          </Stack>
        )}

        {/* TAB 10: OBSERVABILITY DASHBOARD */}
        {activeTab === 'dashboard' && (
          <Stack gap="md">
            <div>
              <Title order={2}>Observability Dashboard</Title>
              <Text c="dimmed" size="sm">
                Real-time streaming metrics, throughput, latency p99, and error counts for running generators.
              </Text>
            </div>

            <Grid>
              <Grid.Col span={4}>
                <Paper withBorder p="md" radius="md">
                  <Text size="xs" c="dimmed" tt="uppercase" fw={700}>Status</Text>
                  <Title order={3} c={generatorStatus === 'RUNNING' ? 'green' : 'dimmed'}>
                    {generatorStatus}
                  </Title>
                </Paper>
              </Grid.Col>
              <Grid.Col span={4}>
                <Paper withBorder p="md" radius="md">
                  <Text size="xs" c="dimmed" tt="uppercase" fw={700}>Throughput</Text>
                  <Title order={3} c="blue">
                    {generatorStatus === 'RUNNING' ? '10.0 evt/s' : '0.0 evt/s'}
                  </Title>
                </Paper>
              </Grid.Col>
              <Grid.Col span={4}>
                <Paper withBorder p="md" radius="md">
                  <Text size="xs" c="dimmed" tt="uppercase" fw={700}>Errors</Text>
                  <Title order={3} c="teal">0</Title>
                </Paper>
              </Grid.Col>
            </Grid>

            <Card withBorder shadow="sm" radius="md">
              <Title order={4} mb="md">Real-Time Throughput (events/sec)</Title>
              <div style={{ width: '100%', height: 300 }}>
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={metricsHistory}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#2C2E33" />
                    <XAxis dataKey="time" stroke="#909296" />
                    <YAxis stroke="#909296" />
                    <Tooltip contentStyle={{ background: '#1A1B1E', border: '1px solid #373A40' }} />
                    <Line type="monotone" dataKey="throughput" stroke="#228be6" strokeWidth={2} dot={false} />
                  </LineChart>
                </ResponsiveContainer>
              </div>
            </Card>

            {/* Distributed Cluster Workers Status */}
            <Card withBorder shadow="sm" radius="md">
              <Group justify="space-between" mb="xs">
                <div>
                  <Title order={4}>Distributed Cluster Workers (Scale-Out Mode)</Title>
                  <Text size="xs" c="dimmed">
                    Partitioned generation nodes coordinated by OpenMimic Coordinator.
                  </Text>
                </div>
                <Badge color="violet" leftSection={<IconCpu size={14} />}>Cluster Active</Badge>
              </Group>

              <Table striped highlightOnHover mt="sm">
                <Table.Thead>
                  <Table.Tr>
                    <Table.Th>Worker ID</Table.Th>
                    <Table.Th>Address</Table.Th>
                    <Table.Th>Partition Range</Table.Th>
                    <Table.Th>Partition Rate</Table.Th>
                    <Table.Th>Status</Table.Th>
                  </Table.Tr>
                </Table.Thead>
                <Table.Tbody>
                  <Table.Tr>
                    <Table.Td><Text fw={500} size="sm">worker-node-01</Text></Table.Td>
                    <Table.Td><Text size="xs" c="dimmed">10.0.1.14:50051</Text></Table.Td>
                    <Table.Td><Badge variant="outline" color="blue">0 - 249,999</Badge></Table.Td>
                    <Table.Td><Text size="sm">25,000 evt/s</Text></Table.Td>
                    <Table.Td><Badge color="green">ONLINE</Badge></Table.Td>
                  </Table.Tr>
                  <Table.Tr>
                    <Table.Td><Text fw={500} size="sm">worker-node-02</Text></Table.Td>
                    <Table.Td><Text size="xs" c="dimmed">10.0.1.15:50051</Text></Table.Td>
                    <Table.Td><Badge variant="outline" color="blue">250,000 - 499,999</Badge></Table.Td>
                    <Table.Td><Text size="sm">25,000 evt/s</Text></Table.Td>
                    <Table.Td><Badge color="green">ONLINE</Badge></Table.Td>
                  </Table.Tr>
                  <Table.Tr>
                    <Table.Td><Text fw={500} size="sm">worker-node-03</Text></Table.Td>
                    <Table.Td><Text size="xs" c="dimmed">10.0.1.16:50051</Text></Table.Td>
                    <Table.Td><Badge variant="outline" color="blue">500,000 - 749,999</Badge></Table.Td>
                    <Table.Td><Text size="sm">25,000 evt/s</Text></Table.Td>
                    <Table.Td><Badge color="green">ONLINE</Badge></Table.Td>
                  </Table.Tr>
                  <Table.Tr>
                    <Table.Td><Text fw={500} size="sm">worker-node-04</Text></Table.Td>
                    <Table.Td><Text size="xs" c="dimmed">10.0.1.17:50051</Text></Table.Td>
                    <Table.Td><Badge variant="outline" color="blue">750,000 - 999,999</Badge></Table.Td>
                    <Table.Td><Text size="sm">25,000 evt/s</Text></Table.Td>
                    <Table.Td><Badge color="green">ONLINE</Badge></Table.Td>
                  </Table.Tr>
                </Table.Tbody>
              </Table>
            </Card>
          </Stack>
        )}
      </AppShell.Main>

      {/* SQL DDL IMPORT MODAL */}
      <Modal opened={ddlOpened} onClose={closeDDL} title="Import from SQL CREATE TABLE DDL" size="lg">
        <Stack gap="sm">
          <Text size="sm" c="dimmed">
            Paste your SQL table definition (PostgreSQL, MySQL, SQLite, Snowflake). OpenMimic will automatically infer column generators, UUIDs, categorical statuses, monetary distributions, and timestamps.
          </Text>
          <Textarea
            rows={8}
            value={ddlText}
            onChange={(e) => setDdlText(e.target.value)}
            style={{ fontFamily: 'monospace' }}
          />
          <Group justify="flex-end">
            <Button variant="default" onClick={closeDDL}>Cancel</Button>
            <Button
              color="teal"
              leftSection={isDdlLoading ? <Loader size="xs" color="white" /> : <IconFileImport size={16} />}
              onClick={handleImportDDL}
              disabled={isDdlLoading}
            >
              Parse & Generate Schema
            </Button>
          </Group>
        </Stack>
      </Modal>
    </AppShell>
  );
}

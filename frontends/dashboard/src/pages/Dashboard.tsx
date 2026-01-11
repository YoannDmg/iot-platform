import { Title, Text, SimpleGrid, Card, Group } from '@mantine/core';
import { IconDevices, IconPlugConnected, IconPlugConnectedX, IconAlertCircle } from '@tabler/icons-react';

export function Dashboard() {
  // TODO: Fetch real stats from GraphQL
  const stats = [
    { title: 'Total Devices', value: '3', icon: IconDevices, color: 'blue' },
    { title: 'Online', value: '1', icon: IconPlugConnected, color: 'green' },
    { title: 'Offline', value: '2', icon: IconPlugConnectedX, color: 'red' },
    { title: 'Errors', value: '0', icon: IconAlertCircle, color: 'orange' },
  ];

  return (
    <>
      <Title order={1} mb="lg">
        Dashboard
      </Title>
      <Text c="dimmed" mb="xl">
        Vue d'ensemble de votre plateforme IoT
      </Text>

      <SimpleGrid cols={{ base: 1, sm: 2, lg: 4 }} spacing="lg">
        {stats.map((stat) => (
          <Card key={stat.title} shadow="sm" padding="lg" radius="md" withBorder>
            <Group justify="space-between" mb="xs">
              <Text fw={500} size="sm">
                {stat.title}
              </Text>
              <stat.icon size={24} color={`var(--mantine-color-${stat.color}-6)`} />
            </Group>
            <Text size="xl" fw={700} c={stat.color}>
              {stat.value}
            </Text>
          </Card>
        ))}
      </SimpleGrid>
    </>
  );
}

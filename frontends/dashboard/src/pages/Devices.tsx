import { Title, Text, Card } from '@mantine/core';

export function Devices() {
  return (
    <>
      <Title order={1} mb="lg">
        Devices
      </Title>
      <Text c="dimmed" mb="xl">
        Liste de tous vos appareils IoT
      </Text>

      <Card shadow="sm" padding="lg" radius="md" withBorder>
        <Text>La liste des devices sera affichée ici</Text>
      </Card>
    </>
  );
}

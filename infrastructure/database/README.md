# Database Infrastructure

Base de données centralisée PostgreSQL pour la plateforme IoT.

## Structure

```
iot_platform (database)
├── devices          # Tables pour device-manager
└── users (à venir)  # Tables pour user-service
```

## Migrations

Toutes les migrations SQL sont centralisées ici:

- `001_create_devices_table.sql` - Création de la table devices

## Usage

Les services pointent vers ces migrations via leur `sqlc.yaml`:

```yaml
schema: "../../infrastructure/database/migrations"
```

## Appliquer les migrations

```bash
make db-migrate
```

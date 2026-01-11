---
id: git-commit-convention
title: Git Commit Convention
sidebar_position: 1
---

# Git Commit Convention

Une convention de commit standardise les messages pour rendre l’historique clair et exploitable. La plus utilisée est La plus utilisée est [Conventional Commits](https://www.conventionalcommits.org).

## Structure d’un commit

```markdown
<type>[scope optionnel]: <sujet en impératif>
[description facultative]
[footer facultatif]
```

* **type** : catégorie du commit (`feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `chore`)
* **scope** *(optionnel)* : partie du projet affectée (ex: `auth`, `api`)
* **sujet** : résumé concis, à l’infinitif, sans point final
* **description** *(optionnelle)* : détails sur le quoi et pourquoi
* **footer** *(optionnel)* : référence d’issue ou breaking change

## Types de commit

```markdown
| Type      | Usage |
|-----------|-------|
| feat      | Nouvelle fonctionnalité |
| fix       | Correction de bug |
| docs      | Documentation |
| style     | Formatage / indentation / lint (pas de modification fonctionnelle) |
| refactor  | Refactoring sans ajout ni correction |
| perf      | Optimisation de performance |
| test      | Ajout ou correction de tests |
| chore     | Tâches annexes (build, config, scripts) |
```

## Exemples

```markdown
feat(auth): add Google login
fix(api): handle null response
docs(readme): update installation instructions
style(ui): fix button alignment
refactor(cart): simplify checkout logic
perf(database): optimize query performance
test(auth): add unit tests for login
chore(build): update dependencies
```

## Footer et références

```markdown
fix(auth): handle invalid token

This fixes the crash when token is expired.

Closes #123
BREAKING CHANGE: login API now requires token version 2
```

* `Closes #123` → ferme automatiquement l’issue liée
* `BREAKING CHANGE` → indique un changement incompatible

## Bonnes pratiques

* Un commit = un changement logique
* Éviter les messages vagues comme “update” ou “fix stuff”
* Commits courts et clairs
* Respecter la convention pour faciliter les outils automatisés (changelog, semantic versioning)

## Ressources

- [Conventional Commits](https://www.conventionalcommits.org/)
- [Git Documentation](https://git-scm.com/doc)
- [Semantic Release](https://semantic-release.gitbook.io/semantic-release/)


import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  // Documentation principale
  docsSidebar: [
    'documentation/overiew',
    'documentation/getting-started',
    'documentation/device-manager',
    'documentation/api-gateway',
  ],

  // Notes d'apprentissage (séparées)
  learningSidebar: [
    'learning/git-commit-convention',
    'learning/LEARNING_NOTES',
    'learning/LEARNING_NOTESV2',
  ],
};

export default sidebars;

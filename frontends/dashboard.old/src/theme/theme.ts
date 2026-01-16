import { createTheme } from '@mantine/core';
import type { MantineColorsTuple } from '@mantine/core';

/**
 * ============================================================================
 * FICHIER DE CONFIGURATION UNIQUE DU STYLE - IoT Platform
 * ============================================================================
 *
 * ✨ CE FICHIER EST LE SEUL ENDROIT OÙ VOUS DEVEZ MODIFIER LES STYLES ✨
 *
 * Vous pouvez ici :
 * - Modifier les couleurs de toute l'application
 * - Changer la typographie (fonts, tailles)
 * - Ajuster les espacements, rayons, ombres
 * - Personnaliser le style de TOUS les composants Mantine
 * - Définir des styles globaux personnalisés
 *
 * Tous les composants et pages utiliseront automatiquement ces styles.
 * ============================================================================
 */

// ============================================================================
// 1. PALETTES DE COULEURS
// ============================================================================
// Modifiez ces palettes pour changer les couleurs de toute l'application
// Chaque palette contient 10 nuances (0=plus clair, 9=plus foncé)

const primaryColor: MantineColorsTuple = [
  '#e6f4ff',
  '#bae0ff',
  '#91caff',
  '#69b1ff',
  '#4096ff',
  '#1677ff', // [5] Couleur principale utilisée par défaut
  '#0958d9',
  '#003eb3',
  '#002c8c',
  '#001d66',
];

const successColor: MantineColorsTuple = [
  '#f6ffed',
  '#d9f7be',
  '#b7eb8f',
  '#95de64',
  '#73d13d',
  '#52c41a', // [5] Couleur de succès
  '#389e0d',
  '#237804',
  '#135200',
  '#092b00',
];

const warningColor: MantineColorsTuple = [
  '#fffbe6',
  '#fff1b8',
  '#ffe58f',
  '#ffd666',
  '#ffc53d',
  '#faad14', // [5] Couleur d'avertissement
  '#d48806',
  '#ad6800',
  '#874d00',
  '#613400',
];

const errorColor: MantineColorsTuple = [
  '#fff1f0',
  '#ffccc7',
  '#ffa39e',
  '#ff7875',
  '#ff4d4f',
  '#f5222d', // [5] Couleur d'erreur
  '#cf1322',
  '#a8071a',
  '#820014',
  '#5c0011',
];

// ============================================================================
// 2. CONFIGURATION DU THÈME
// ============================================================================

export const theme = createTheme({
  // ---------------------------------------------------------------------------
  // COULEURS
  // ---------------------------------------------------------------------------
  colors: {
    primary: primaryColor,
    success: successColor,
    warning: warningColor,
    error: errorColor,
  },
  primaryColor: 'primary',
  primaryShade: 5, // Utilise la nuance [5] par défaut

  // ---------------------------------------------------------------------------
  // RAYON DES BORDURES (Border Radius)
  // ---------------------------------------------------------------------------
  defaultRadius: 'md',
  radius: {
    xs: '4px',
    sm: '6px',
    md: '8px',
    lg: '12px',
    xl: '16px',
  },

  // ---------------------------------------------------------------------------
  // ESPACEMENTS (Margin, Padding, Gap)
  // ---------------------------------------------------------------------------
  spacing: {
    xs: '8px',
    sm: '12px',
    md: '16px',
    lg: '24px',
    xl: '32px',
  },

  // ---------------------------------------------------------------------------
  // TYPOGRAPHIE
  // ---------------------------------------------------------------------------
  fontFamily: 'Roboto, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", "Helvetica Neue", Arial, sans-serif',
  fontFamilyMonospace: 'Roboto, ui-monospace, SFMono-Regular, "SF Mono", Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace',

  //fontFamily: '"JetBrains Mono", monospace',
  //fontFamilyMonospace: '"JetBrains Mono", monospace',

  // Tailles des titres
  headings: {
    fontFamily: 'Roboto, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", "Helvetica Neue", Arial, sans-serif',
    //fontFamily: '"JetBrains Mono", monospace',
    fontWeight: '600',
    sizes: {
      h1: { fontSize: '2rem', lineHeight: '1.2' },
      h2: { fontSize: '1.75rem', lineHeight: '1.25' },
      h3: { fontSize: '1.5rem', lineHeight: '1.3' },
      h4: { fontSize: '1.25rem', lineHeight: '1.35' },
      h5: { fontSize: '1.125rem', lineHeight: '1.4' },
      h6: { fontSize: '1rem', lineHeight: '1.45' },
    },
  },

  // ---------------------------------------------------------------------------
  // OMBRES (Box Shadows)
  // ---------------------------------------------------------------------------
  shadows: {
    xs: '0 1px 2px rgba(0, 0, 0, 0.05)',
    sm: '0 1px 3px rgba(0, 0, 0, 0.1), 0 1px 2px rgba(0, 0, 0, 0.06)',
    md: '0 4px 6px rgba(0, 0, 0, 0.1), 0 2px 4px rgba(0, 0, 0, 0.06)',
    lg: '0 10px 15px rgba(0, 0, 0, 0.1), 0 4px 6px rgba(0, 0, 0, 0.05)',
    xl: '0 20px 25px rgba(0, 0, 0, 0.1), 0 10px 10px rgba(0, 0, 0, 0.04)',
  },

  // ---------------------------------------------------------------------------
  // PERSONNALISATION DES COMPOSANTS
  // ---------------------------------------------------------------------------
  // Vous pouvez personnaliser TOUS les composants Mantine ici
  // Les modifications s'appliquent automatiquement à toute l'application

  components: {
    Button: {
      defaultProps: {
        radius: 'md',
      },
      // Exemple : ajouter des styles custom
      // styles: {
      //   root: {
      //     fontWeight: 500,
      //   },
      // },
    },

    Card: {
      defaultProps: {
        padding: 'lg',
        radius: 'md',
        withBorder: true,
      },
      // styles: {
      //   root: {
      //     transition: 'transform 0.2s',
      //     '&:hover': {
      //       transform: 'translateY(-2px)',
      //     },
      //   },
      // },
    },

    Paper: {
      defaultProps: {
        padding: 'md',
        radius: 'md',
      },
    },

    TextInput: {
      defaultProps: {
        radius: 'md',
      },
    },

    Select: {
      defaultProps: {
        radius: 'md',
      },
    },

    Modal: {
      defaultProps: {
        radius: 'md',
        centered: true,
      },
    },

    Notification: {
      defaultProps: {
        radius: 'md',
      },
    },

    AppShell: {
      styles: {
        main: {
          // Style du contenu principal
        },
      },
    },

    NavLink: {
      styles: {
        root: {
          borderRadius: '8px',
          // Ajoutez vos styles personnalisés pour les liens de navigation
        },
      },
    },

    // Ajoutez d'autres composants ici selon vos besoins
    // Table, Badge, Tabs, Input, etc.
  },

  // ---------------------------------------------------------------------------
  // AUTRES CONFIGURATIONS
  // ---------------------------------------------------------------------------

  // Breakpoints responsive (facultatif, personnalisable)
  // breakpoints: {
  //   xs: '36em',  // 576px
  //   sm: '48em',  // 768px
  //   md: '62em',  // 992px
  //   lg: '75em',  // 1200px
  //   xl: '88em',  // 1408px
  // },

  // Couleurs supplémentaires (facultatif)
  // white: '#ffffff',
  // black: '#000000',
});

/**
 * ============================================================================
 * COMMENT UTILISER CE FICHIER
 * ============================================================================
 *
 * 1. Pour changer une couleur globale :
 *    Modifiez les valeurs hexadécimales dans les palettes de couleurs
 *
 * 2. Pour changer la police :
 *    Modifiez `fontFamily` et `headings.fontFamily`
 *
 * 3. Pour modifier un composant spécifique :
 *    Ajoutez ou modifiez son entrée dans `components`
 *
 * 4. Pour ajouter des styles globaux :
 *    Utilisez la propriété `styles` dans la config du composant
 *
 * Exemple de personnalisation d'un composant :
 *
 * Button: {
 *   defaultProps: {
 *     size: 'md',
 *     variant: 'filled',
 *   },
 *   styles: (theme) => ({
 *     root: {
 *       fontWeight: 600,
 *       '&:hover': {
 *         transform: 'scale(1.05)',
 *       },
 *     },
 *   }),
 * },
 *
 * Tous les composants de votre application utiliseront ces styles !
 * ============================================================================
 */

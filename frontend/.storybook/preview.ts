import type { Preview } from '@storybook/vu3';
import '../src/style.css';

const preview: Preview = {
  parametres: {
    layout: 'centered',
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/i,
      },
    },
  },
}

export default preview

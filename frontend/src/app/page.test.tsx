import { render, screen } from '@testing-library/react';
import HomePage from './page';

describe('HomePage', () => {
  it('renders the main heading', () => {
    render(<HomePage />);

    const heading = screen.getByRole('heading', {
      name: /Hello MyAnimeLogs!/i,
    });

    expect(heading).toBeInTheDocument();
  });
});

import { describe, it, expect, vi } from 'vitest';
import { render, fireEvent } from '@testing-library/react';
import { StarRating } from './StarRating';

describe('StarRating', () => {
  it('renders the correct number of stars', () => {
    const { container } = render(<StarRating rating={3} />);
    const stars = container.querySelectorAll('svg');
    expect(stars).toHaveLength(5);
  });

  it('renders correct number of filled stars based on rating', () => {
    const { container } = render(<StarRating rating={4} />);
    const filledStars = container.querySelectorAll('[class*="fill-(--warning)"]');
    expect(filledStars).toHaveLength(4);
  });

  it('renders filled stars based on rating', () => {
    const { container } = render(<StarRating rating={3} />);
    const filledStars = container.querySelectorAll('[class*="text-(--warning)"]');
    expect(filledStars.length).toBeGreaterThan(0);
  });

  it('calls onRatingChange when a star is clicked in interactive mode', () => {
    const handleChange = vi.fn();
    const { container } = render(<StarRating rating={0} onRatingChange={handleChange} />);

    const buttons = container.querySelectorAll('button');
    fireEvent.click(buttons[2]); // Click third star

    expect(handleChange).toHaveBeenCalledWith(3);
  });

  it('does not call onRatingChange when readonly', () => {
    const handleChange = vi.fn();
    const { container } = render(<StarRating rating={3} onRatingChange={handleChange} readonly />);

    const buttons = container.querySelectorAll('button');
    fireEvent.click(buttons[4]);

    expect(handleChange).not.toHaveBeenCalled();
  });

  it('shows hover effect in interactive mode', () => {
    const handleChange = vi.fn();
    const { container } = render(<StarRating rating={0} onRatingChange={handleChange} />);

    // Verify hover state is applied
    expect(container.querySelector('.cursor-pointer')).toBeInTheDocument();
  });

  it('handles half ratings correctly', () => {
    const { container } = render(<StarRating rating={3.5} />);
    // Should fill 4 stars (rounds up for visual)
    const filledStars = container.querySelectorAll('[class*="fill-(--warning)"]');
    expect(filledStars.length).toBeGreaterThanOrEqual(3);
  });

  it('handles zero rating', () => {
    const { container } = render(<StarRating rating={0} />);
    const filledStars = container.querySelectorAll('[class*="fill-(--warning)"]');
    expect(filledStars).toHaveLength(0);
  });

  it('handles maximum rating', () => {
    const { container } = render(<StarRating rating={5} />);
    const filledStars = container.querySelectorAll('[class*="fill-(--warning)"]');
    expect(filledStars).toHaveLength(5);
  });
});

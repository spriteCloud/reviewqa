import { useState } from 'react';

interface CounterProps {
  initial?: number;
  step?: number;
}

/**
 * Small interactive counter used to demo reviewqa's multi-scenario Playwright
 * generation. Has a numeric display and an increment button with stable
 * test-ids, plus a useState hook so the click-toggles-state scenario kicks in.
 */
export function Counter({ initial = 0, step = 1 }: CounterProps) {
  const [value, setValue] = useState(initial);

  return (
    <div data-testid="counter-root" role="region" aria-label="Counter widget">
      <span data-testid="counter-display">{value}</span>
      <button
        type="button"
        data-testid="counter-inc"
        aria-expanded={value > 0}
        onClick={() => setValue((v) => v + step)}
      >
        +
      </button>
    </div>
  );
}

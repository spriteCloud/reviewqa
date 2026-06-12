import { useState } from 'react';

interface QA {
  question: string;
  answer: string;
}

interface FAQProps {
  items: QA[];
}

/**
 * Accordion FAQ block used on the landing page. Each item expands one at a
 * time. The static HTML version on `index.html` uses `<details>` elements;
 * this component is the future client-side replacement once the page
 * hydrates.
 */
export function FAQ({ items }: FAQProps) {
  const [openIndex, setOpenIndex] = useState<number | null>(null);

  return (
    <div className="faq-list" data-testid="faq-list">
      {items.map((item, i) => {
        const isOpen = openIndex === i;
        return (
          <div
            key={item.question}
            className={`faq-item${isOpen ? ' is-open' : ''}`}
            data-testid="faq-item"
          >
            <button
              type="button"
              data-testid="faq-summary"
              aria-expanded={isOpen}
              onClick={() => setOpenIndex(isOpen ? null : i)}
            >
              <span>{item.question}</span>
              <span aria-hidden="true">{isOpen ? '−' : '+'}</span>
            </button>
            {isOpen && (
              <p data-testid="faq-answer">{item.answer}</p>
            )}
          </div>
        );
      })}
    </div>
  );
}

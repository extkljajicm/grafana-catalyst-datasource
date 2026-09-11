// HelpTooltip.tsx: Reusable tooltip component for providing inline help
import React from 'react';
import { Tooltip, Icon } from '@grafana/ui';

export interface HelpTooltipProps {
  content: string | React.ReactNode;
  /** Optional link to documentation */
  link?: string;
}

/**
 * HelpTooltip displays a help icon with tooltip content
 * Useful for providing contextual help for form fields
 */
export const HelpTooltip: React.FC<HelpTooltipProps> = ({ content, link }) => {
  const tooltipContent = (
    <div>
      {content}
      {link && (
        <div style={{ marginTop: '8px' }}>
          <a href={link} target="_blank" rel="noopener noreferrer">
            Learn more →
          </a>
        </div>
      )}
    </div>
  );

  return (
    <Tooltip content={tooltipContent} placement="right">
      <Icon name="info-circle" size="sm" style={{ marginLeft: '4px', cursor: 'help' }} />
    </Tooltip>
  );
};

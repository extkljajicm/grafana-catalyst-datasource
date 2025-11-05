// ConfigEditor.test.tsx: Unit tests for ConfigEditor component
import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import { ConfigEditor } from './ConfigEditor';
import type { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import type { CatalystJsonData } from '../types';

// Mock validation module
jest.mock('../validation', () => ({
  validateConfig: jest.fn((config) => {
    const errors: string[] = [];
    if (!config.baseUrl || config.baseUrl.trim() === '') {
      errors.push('Base URL is required');
    } else {
      try {
        const url = new URL(config.baseUrl);
        if (!url.protocol.startsWith('http')) {
          errors.push('Base URL must use HTTP or HTTPS protocol');
        }
      } catch {
        errors.push('Base URL is not a valid URL');
      }
    }
    return {
      valid: errors.length === 0,
      errors,
    };
  }),
}));

// Mock HelpTooltip component
jest.mock('./HelpTooltip', () => ({
  HelpTooltip: () => null,
}));

type SecureShape = {
  username?: string;
  password?: string;
  apiToken?: string;
};

type Props = DataSourcePluginOptionsEditorProps<CatalystJsonData, SecureShape>;

describe('ConfigEditor', () => {
  let mockOnOptionsChange: jest.Mock;
  let defaultProps: Props;

  beforeEach(() => {
    mockOnOptionsChange = jest.fn();

    defaultProps = {
      options: {
        id: 1,
        uid: 'test-uid',
        orgId: 1,
        name: 'Test Catalyst',
        type: 'catalyst-datasource',
        typeName: 'Catalyst',
        typeLogoUrl: '',
        access: 'proxy',
        url: '',
        user: '',
        database: '',
        basicAuth: false,
        basicAuthUser: '',
        withCredentials: false,
        isDefault: false,
        jsonData: {},
        secureJsonFields: {},
        version: 1,
        readOnly: false,
      },
      onOptionsChange: mockOnOptionsChange,
    } as any;
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render without crashing', () => {
      render(<ConfigEditor {...defaultProps} />);

      expect(screen.getByText('API Endpoint')).toBeInTheDocument();
      expect(screen.getByText('Catalyst Base URL')).toBeInTheDocument();
      expect(screen.getByText('Skip TLS verification')).toBeInTheDocument();
      expect(screen.getByText('Username')).toBeInTheDocument();
      expect(screen.getByText('Password')).toBeInTheDocument();
      expect(screen.getByText('API Token (override)')).toBeInTheDocument();
    });

    it('should render all input fields', () => {
      render(<ConfigEditor {...defaultProps} />);

      // Check for Base URL input
      const baseUrlInput = screen.getByPlaceholderText('https://catalyst.example.com');
      expect(baseUrlInput).toBeInTheDocument();

      // Check for Username input
      const usernameInput = screen.getByPlaceholderText('dnac-api-user');
      expect(usernameInput).toBeInTheDocument();
    });

    it('should display existing base URL', () => {
      const props = {
        ...defaultProps,
        options: {
          ...defaultProps.options,
          jsonData: {
            baseUrl: 'https://catalyst.test.com',
          },
        },
      };

      render(<ConfigEditor {...props} />);

      const baseUrlInput = screen.getByPlaceholderText('https://catalyst.example.com') as HTMLInputElement;
      expect(baseUrlInput.value).toBe('https://catalyst.test.com');
    });

    it('should display existing endpoint selection', () => {
      const props = {
        ...defaultProps,
        options: {
          ...defaultProps.options,
          jsonData: {
            endpoint: 'siteHealth',
          },
        },
      };

      render(<ConfigEditor {...props} />);

      expect(screen.getByText('Site Health')).toBeInTheDocument();
    });

    it('should default endpoint to alerts when not set', () => {
      render(<ConfigEditor {...defaultProps} />);

      expect(screen.getByText('Issues/Alerts')).toBeInTheDocument();
    });

    it('should display TLS skip verification switch', () => {
      render(<ConfigEditor {...defaultProps} />);

      const tlsLabel = screen.getByText('Skip TLS verification');
      const fieldContainer = tlsLabel.closest('.css-1rplq84');
      const tlsSwitch = fieldContainer?.querySelector('input[type="checkbox"]') as HTMLInputElement;

      expect(tlsSwitch).toBeTruthy();
      expect(tlsSwitch.checked).toBe(false);
    });

    it('should display TLS skip verification as checked when set', () => {
      const props = {
        ...defaultProps,
        options: {
          ...defaultProps.options,
          jsonData: {
            insecureSkipVerify: true,
          },
        },
      };

      render(<ConfigEditor {...props} />);

      const tlsLabel = screen.getByText('Skip TLS verification');
      const fieldContainer = tlsLabel.closest('.css-1rplq84');
      const tlsSwitch = fieldContainer?.querySelector('input[type="checkbox"]') as HTMLInputElement;

      expect(tlsSwitch.checked).toBe(true);
    });
  });

  describe('Input Handling', () => {
    it('should update Base URL when typing', () => {
      render(<ConfigEditor {...defaultProps} />);

      const baseUrlInput = screen.getByPlaceholderText('https://catalyst.example.com');
      fireEvent.change(baseUrlInput, { target: { value: 'https://new-url.com' } });

      expect(mockOnOptionsChange).toHaveBeenCalledWith(
        expect.objectContaining({
          jsonData: expect.objectContaining({
            baseUrl: 'https://new-url.com',
          }),
        })
      );
    });

    it('should update username when typing', () => {
      render(<ConfigEditor {...defaultProps} />);

      const usernameInput = screen.getByPlaceholderText('dnac-api-user');
      fireEvent.change(usernameInput, { target: { value: 'test-user' } });

      expect(mockOnOptionsChange).toHaveBeenCalledWith(
        expect.objectContaining({
          secureJsonData: expect.objectContaining({
            username: 'test-user',
          }),
        })
      );
    });

    it('should preserve existing jsonData when updating Base URL', () => {
      const props = {
        ...defaultProps,
        options: {
          ...defaultProps.options,
          jsonData: {
            endpoint: 'siteHealth',
            baseUrl: 'https://old-url.com',
          },
        },
      };

      render(<ConfigEditor {...props} />);

      const baseUrlInput = screen.getByPlaceholderText('https://catalyst.example.com');
      fireEvent.change(baseUrlInput, { target: { value: 'https://new-url.com' } });

      expect(mockOnOptionsChange).toHaveBeenCalledWith(
        expect.objectContaining({
          jsonData: expect.objectContaining({
            baseUrl: 'https://new-url.com',
            endpoint: 'siteHealth',
          }),
        })
      );
    });

    it('should toggle TLS skip verification', () => {
      render(<ConfigEditor {...defaultProps} />);

      const tlsLabel = screen.getByText('Skip TLS verification');
      const fieldContainer = tlsLabel.closest('.css-1rplq84');
      const tlsSwitch = fieldContainer?.querySelector('input[type="checkbox"]') as HTMLInputElement;

      fireEvent.click(tlsSwitch);

      expect(mockOnOptionsChange).toHaveBeenCalledWith(
        expect.objectContaining({
          jsonData: expect.objectContaining({
            insecureSkipVerify: true,
          }),
        })
      );
    });

    it('should toggle TLS skip verification off when already on', () => {
      const props = {
        ...defaultProps,
        options: {
          ...defaultProps.options,
          jsonData: {
            insecureSkipVerify: true,
          },
        },
      };

      render(<ConfigEditor {...props} />);

      const tlsLabel = screen.getByText('Skip TLS verification');
      const fieldContainer = tlsLabel.closest('.css-1rplq84');
      const tlsSwitch = fieldContainer?.querySelector('input[type="checkbox"]') as HTMLInputElement;

      fireEvent.click(tlsSwitch);

      expect(mockOnOptionsChange).toHaveBeenCalledWith(
        expect.objectContaining({
          jsonData: expect.objectContaining({
            insecureSkipVerify: false,
          }),
        })
      );
    });
  });

  describe('Secure Fields', () => {
    it('should update password when typing', () => {
      render(<ConfigEditor {...defaultProps} />);

      const passwordInputs = screen.getAllByPlaceholderText(/••••••••/);
      const passwordInput = passwordInputs[0]; // First one is password, second is API token

      fireEvent.change(passwordInput, { target: { value: 'new-password' } });

      expect(mockOnOptionsChange).toHaveBeenCalledWith(
        expect.objectContaining({
          secureJsonData: expect.objectContaining({
            password: 'new-password',
          }),
        })
      );
    });

    it('should update API token when typing', () => {
      render(<ConfigEditor {...defaultProps} />);

      const tokenInput = screen.getByPlaceholderText('(optional) Paste X-Auth-Token');
      fireEvent.change(tokenInput, { target: { value: 'test-token-123' } });

      expect(mockOnOptionsChange).toHaveBeenCalledWith(
        expect.objectContaining({
          secureJsonData: expect.objectContaining({
            apiToken: 'test-token-123',
          }),
        })
      );
    });

    it('should show password as configured when secureJsonFields.password is true', () => {
      const props = {
        ...defaultProps,
        options: {
          ...defaultProps.options,
          secureJsonFields: {
            password: true,
          },
        },
      };

      render(<ConfigEditor {...props} />);

      // When password is configured, the Reset button should be visible
      const resetButtons = screen.getAllByText('Reset');
      expect(resetButtons.length).toBeGreaterThan(0);
    });

    it('should show API token as configured when secureJsonFields.apiToken is true', () => {
      const props = {
        ...defaultProps,
        options: {
          ...defaultProps.options,
          secureJsonFields: {
            apiToken: true,
          },
        },
      };

      render(<ConfigEditor {...props} />);

      // SecretInput shows Reset button for configured fields
      const resetButtons = screen.getAllByText('Reset');
      expect(resetButtons.length).toBeGreaterThan(0);
    });

    it('should preserve existing secureJsonData when updating password', () => {
      const props = {
        ...defaultProps,
        options: {
          ...defaultProps.options,
          secureJsonData: {
            username: 'existing-user',
          },
        },
      };

      render(<ConfigEditor {...props} />);

      const passwordInputs = screen.getAllByPlaceholderText(/••••••••/);
      const passwordInput = passwordInputs[0];

      fireEvent.change(passwordInput, { target: { value: 'new-password' } });

      expect(mockOnOptionsChange).toHaveBeenCalledWith(
        expect.objectContaining({
          secureJsonData: expect.objectContaining({
            username: 'existing-user',
            password: 'new-password',
          }),
        })
      );
    });
  });

  describe('Secure Field Reset', () => {
    it('should reset password when clicking reset button', () => {
      const props = {
        ...defaultProps,
        options: {
          ...defaultProps.options,
          secureJsonFields: {
            password: true,
          },
        },
      };

      render(<ConfigEditor {...props} />);

      const resetButtons = screen.getAllByText('Reset');
      const passwordResetButton = resetButtons[0]; // First reset button is for password

      fireEvent.click(passwordResetButton);

      expect(mockOnOptionsChange).toHaveBeenCalledWith(
        expect.objectContaining({
          secureJsonFields: expect.objectContaining({
            password: false,
          }),
          secureJsonData: expect.objectContaining({
            password: '',
          }),
        })
      );
    });

    it('should reset API token when clicking reset button', () => {
      const props = {
        ...defaultProps,
        options: {
          ...defaultProps.options,
          secureJsonFields: {
            apiToken: true,
          },
        },
      };

      render(<ConfigEditor {...props} />);

      const resetButtons = screen.getAllByText('Reset');
      const tokenResetButton = resetButtons[resetButtons.length - 1]; // Last reset button is for API token

      fireEvent.click(tokenResetButton);

      expect(mockOnOptionsChange).toHaveBeenCalledWith(
        expect.objectContaining({
          secureJsonFields: expect.objectContaining({
            apiToken: false,
          }),
          secureJsonData: expect.objectContaining({
            apiToken: '',
          }),
        })
      );
    });

    it('should preserve other secure fields when resetting password', () => {
      const props = {
        ...defaultProps,
        options: {
          ...defaultProps.options,
          secureJsonFields: {
            password: true,
            apiToken: true,
          },
          secureJsonData: {
            username: 'test-user',
          },
        },
      };

      render(<ConfigEditor {...props} />);

      const resetButtons = screen.getAllByText('Reset');
      const passwordResetButton = resetButtons[0];

      fireEvent.click(passwordResetButton);

      expect(mockOnOptionsChange).toHaveBeenCalledWith(
        expect.objectContaining({
          secureJsonFields: expect.objectContaining({
            password: false,
          }),
          secureJsonData: expect.objectContaining({
            username: 'test-user',
            password: '',
          }),
        })
      );
    });
  });

  describe('Validation', () => {
    it('should show validation error for invalid Base URL', async () => {
      render(<ConfigEditor {...defaultProps} />);

      const baseUrlInput = screen.getByPlaceholderText('https://catalyst.example.com');
      fireEvent.change(baseUrlInput, { target: { value: 'invalid-url' } });

      await waitFor(() => {
        expect(screen.getByText(/Configuration Error/)).toBeInTheDocument();
        expect(screen.getByText(/Base URL is not a valid URL/)).toBeInTheDocument();
      });
    });

    it('should show validation error for empty Base URL', async () => {
      const props = {
        ...defaultProps,
        options: {
          ...defaultProps.options,
          jsonData: {
            baseUrl: 'https://valid-url.com',
          },
        },
      };

      render(<ConfigEditor {...props} />);

      const baseUrlInput = screen.getByPlaceholderText('https://catalyst.example.com');
      fireEvent.change(baseUrlInput, { target: { value: '' } });

      await waitFor(() => {
        expect(screen.getByText(/Configuration Error/)).toBeInTheDocument();
        expect(screen.getByText(/Base URL is required/)).toBeInTheDocument();
      });
    });

    it('should show validation error for non-HTTP protocol', async () => {
      render(<ConfigEditor {...defaultProps} />);

      const baseUrlInput = screen.getByPlaceholderText('https://catalyst.example.com');
      fireEvent.change(baseUrlInput, { target: { value: 'ftp://catalyst.example.com' } });

      await waitFor(() => {
        expect(screen.getByText(/Configuration Error/)).toBeInTheDocument();
        expect(screen.getByText(/Base URL must use HTTP or HTTPS protocol/)).toBeInTheDocument();
      });
    });

    it('should clear validation error when valid URL is entered', async () => {
      render(<ConfigEditor {...defaultProps} />);

      const baseUrlInput = screen.getByPlaceholderText('https://catalyst.example.com');

      // Enter invalid URL
      fireEvent.change(baseUrlInput, { target: { value: 'invalid-url' } });

      await waitFor(() => {
        expect(screen.getByText(/Configuration Error/)).toBeInTheDocument();
      });

      // Enter valid URL
      fireEvent.change(baseUrlInput, { target: { value: 'https://valid-url.com' } });

      await waitFor(() => {
        expect(screen.queryByText(/Configuration Error/)).not.toBeInTheDocument();
      });
    });

    it('should mark Base URL field as invalid when validation fails', async () => {
      render(<ConfigEditor {...defaultProps} />);

      const baseUrlInput = screen.getByPlaceholderText('https://catalyst.example.com');
      fireEvent.change(baseUrlInput, { target: { value: 'invalid-url' } });

      await waitFor(() => {
        // Check that the validation error appears
        expect(screen.getByText(/Configuration Error/)).toBeInTheDocument();
      });
    });

    it('should show field-level error message for invalid Base URL', async () => {
      render(<ConfigEditor {...defaultProps} />);

      const baseUrlInput = screen.getByPlaceholderText('https://catalyst.example.com');
      fireEvent.change(baseUrlInput, { target: { value: 'invalid-url' } });

      await waitFor(() => {
        expect(screen.getByText('Invalid Base URL format')).toBeInTheDocument();
      });
    });
  });

  describe('Endpoint Selection', () => {
    it('should render endpoint dropdown with correct options', () => {
      render(<ConfigEditor {...defaultProps} />);

      expect(screen.getByText('Issues/Alerts')).toBeInTheDocument();
    });

    it('should display different endpoint when configured', () => {
      // Test with siteHealth endpoint
      const props = {
        ...defaultProps,
        options: {
          ...defaultProps.options,
          jsonData: {
            endpoint: 'siteHealth',
          },
        },
      };

      const { rerender } = render(<ConfigEditor {...props} />);
      expect(screen.getByText('Site Health')).toBeInTheDocument();

      // Test with alerts endpoint
      const alertsProps = {
        ...defaultProps,
        options: {
          ...defaultProps.options,
          jsonData: {
            endpoint: 'alerts',
          },
        },
      };

      rerender(<ConfigEditor {...alertsProps} />);
      expect(screen.getByText('Issues/Alerts')).toBeInTheDocument();
    });
  });

  describe('Edge Cases', () => {
    it('should handle undefined jsonData gracefully', () => {
      const props = {
        ...defaultProps,
        options: {
          ...defaultProps.options,
          jsonData: undefined as any,
        },
      };

      render(<ConfigEditor {...props} />);

      const baseUrlInput = screen.getByPlaceholderText('https://catalyst.example.com') as HTMLInputElement;
      expect(baseUrlInput.value).toBe('');
    });

    it('should handle undefined secureJsonData gracefully', () => {
      const props = {
        ...defaultProps,
        options: {
          ...defaultProps.options,
          secureJsonData: undefined as any,
        },
      };

      render(<ConfigEditor {...props} />);

      const usernameInput = screen.getByPlaceholderText('dnac-api-user') as HTMLInputElement;
      expect(usernameInput.value).toBe('');
    });

    it('should handle undefined secureJsonFields gracefully', () => {
      const props = {
        ...defaultProps,
        options: {
          ...defaultProps.options,
          secureJsonFields: undefined as any,
        },
      };

      render(<ConfigEditor {...props} />);

      // Should render without crashing
      expect(screen.getByText('Password')).toBeInTheDocument();
    });

    it('should handle simultaneous updates to multiple fields', () => {
      render(<ConfigEditor {...defaultProps} />);

      const baseUrlInput = screen.getByPlaceholderText('https://catalyst.example.com');
      const usernameInput = screen.getByPlaceholderText('dnac-api-user');

      fireEvent.change(baseUrlInput, { target: { value: 'https://test.com' } });
      fireEvent.change(usernameInput, { target: { value: 'test-user' } });

      // Both calls should have been made
      expect(mockOnOptionsChange).toHaveBeenCalledTimes(2);
    });
  });
});

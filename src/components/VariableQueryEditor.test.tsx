// VariableQueryEditor.test.tsx: Unit tests for VariableQueryEditor component
import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import VariableQueryEditor from './VariableQueryEditor';
import type { CatalystVariableQuery } from '../types';

describe('VariableQueryEditor', () => {
  let mockOnChange: jest.Mock;
  
  beforeEach(() => {
    mockOnChange = jest.fn();
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render without crashing', () => {
      const query: CatalystVariableQuery = { type: 'priorities' };
      render(<VariableQueryEditor query={query} onChange={mockOnChange} />);
      
      expect(screen.getByText('Type')).toBeInTheDocument();
      expect(screen.getByText('Definition')).toBeInTheDocument();
    });

    it('should render with undefined query', () => {
      render(<VariableQueryEditor query={undefined} onChange={mockOnChange} />);
      
      expect(screen.getByText('Type')).toBeInTheDocument();
      expect(screen.getByText('Definition')).toBeInTheDocument();
    });

    it('should display correct default query type (priorities)', () => {
      const query: CatalystVariableQuery = { type: 'priorities' };
      render(<VariableQueryEditor query={query} onChange={mockOnChange} />);
      
      // Check that "Priorities (P1..P4)" is displayed
      expect(screen.getByText('Priorities (P1..P4)')).toBeInTheDocument();
    });

    it('should not show search field for priorities', () => {
      const query: CatalystVariableQuery = { type: 'priorities' };
      render(<VariableQueryEditor query={query} onChange={mockOnChange} />);
      
      // Search field should not be present
      expect(screen.queryByText('Search (optional)')).not.toBeInTheDocument();
    });

    it('should not show search field for issueStatuses', () => {
      const query: CatalystVariableQuery = { type: 'issueStatuses' };
      render(<VariableQueryEditor query={query} onChange={mockOnChange} />);
      
      // Search field should not be present
      expect(screen.queryByText('Search (optional)')).not.toBeInTheDocument();
    });

    it('should show search field for sites', () => {
      const query: CatalystVariableQuery = { type: 'sites' };
      render(<VariableQueryEditor query={query} onChange={mockOnChange} />);
      
      expect(screen.getByText('Search (optional)')).toBeInTheDocument();
    });

    it('should show search field for devices', () => {
      const query: CatalystVariableQuery = { type: 'devices' };
      render(<VariableQueryEditor query={query} onChange={mockOnChange} />);
      
      expect(screen.getByText('Search (optional)')).toBeInTheDocument();
    });

    it('should show search field for macs', () => {
      const query: CatalystVariableQuery = { type: 'macs' };
      render(<VariableQueryEditor query={query} onChange={mockOnChange} />);
      
      expect(screen.getByText('Search (optional)')).toBeInTheDocument();
    });
  });

  describe('Query Type Selection', () => {
    it('should render different query types correctly', () => {
      // Test priorities
      const prioritiesQuery: CatalystVariableQuery = { type: 'priorities' };
      const { rerender } = render(<VariableQueryEditor query={prioritiesQuery} onChange={mockOnChange} />);
      expect(screen.getByText('Priorities (P1..P4)')).toBeInTheDocument();
      expect(screen.getByDisplayValue('priorities()')).toBeInTheDocument();
      
      // Test issueStatuses
      const statusQuery: CatalystVariableQuery = { type: 'issueStatuses' };
      rerender(<VariableQueryEditor query={statusQuery} onChange={mockOnChange} />);
      expect(screen.getByText('Issue Statuses')).toBeInTheDocument();
      expect(screen.getByDisplayValue('issueStatuses()')).toBeInTheDocument();
      
      // Test sites
      const sitesQuery: CatalystVariableQuery = { type: 'sites' };
      rerender(<VariableQueryEditor query={sitesQuery} onChange={mockOnChange} />);
      expect(screen.getByText('Sites')).toBeInTheDocument();
      expect(screen.getByDisplayValue('sites()')).toBeInTheDocument();
      
      // Test devices
      const devicesQuery: CatalystVariableQuery = { type: 'devices' };
      rerender(<VariableQueryEditor query={devicesQuery} onChange={mockOnChange} />);
      expect(screen.getByText('Devices')).toBeInTheDocument();
      expect(screen.getByDisplayValue('devices()')).toBeInTheDocument();
      
      // Test macs
      const macsQuery: CatalystVariableQuery = { type: 'macs' };
      rerender(<VariableQueryEditor query={macsQuery} onChange={mockOnChange} />);
      expect(screen.getByText('MACs')).toBeInTheDocument();
      expect(screen.getByDisplayValue('macs()')).toBeInTheDocument();
    });
  });

  describe('Different Query Types Rendering', () => {
    it('should display correct fields for priorities query type', () => {
      const query: CatalystVariableQuery = { type: 'priorities' };
      render(<VariableQueryEditor query={query} onChange={mockOnChange} />);
      
      expect(screen.getByText('Priorities (P1..P4)')).toBeInTheDocument();
      expect(screen.queryByText('Search (optional)')).not.toBeInTheDocument();
      
      const definitionInput = screen.getByDisplayValue('priorities()');
      expect(definitionInput).toBeInTheDocument();
      expect(definitionInput).toHaveAttribute('readonly');
    });

    it('should display correct fields for issueStatuses query type', () => {
      const query: CatalystVariableQuery = { type: 'issueStatuses' };
      render(<VariableQueryEditor query={query} onChange={mockOnChange} />);
      
      expect(screen.getByText('Issue Statuses')).toBeInTheDocument();
      expect(screen.queryByText('Search (optional)')).not.toBeInTheDocument();
      
      const definitionInput = screen.getByDisplayValue('issueStatuses()');
      expect(definitionInput).toBeInTheDocument();
    });

    it('should display correct fields for sites query type without search', () => {
      const query: CatalystVariableQuery = { type: 'sites' };
      render(<VariableQueryEditor query={query} onChange={mockOnChange} />);
      
      expect(screen.getByText('Sites')).toBeInTheDocument();
      expect(screen.getByText('Search (optional)')).toBeInTheDocument();
      
      const definitionInput = screen.getByDisplayValue('sites()');
      expect(definitionInput).toBeInTheDocument();
    });

    it('should display correct fields for sites query type with search', () => {
      const query: CatalystVariableQuery = { type: 'sites', search: 'building' };
      render(<VariableQueryEditor query={query} onChange={mockOnChange} />);
      
      expect(screen.getByText('Sites')).toBeInTheDocument();
      expect(screen.getByText('Search (optional)')).toBeInTheDocument();
      
      const searchInput = screen.getByPlaceholderText(/e.g. branch-a/);
      expect(searchInput).toHaveValue('building');
      
      const definitionInput = screen.getByDisplayValue('sites(search:"building")');
      expect(definitionInput).toBeInTheDocument();
    });

    it('should display correct fields for devices query type', () => {
      const query: CatalystVariableQuery = { type: 'devices' };
      render(<VariableQueryEditor query={query} onChange={mockOnChange} />);
      
      expect(screen.getByText('Devices')).toBeInTheDocument();
      expect(screen.getByText('Search (optional)')).toBeInTheDocument();
      
      const definitionInput = screen.getByDisplayValue('devices()');
      expect(definitionInput).toBeInTheDocument();
    });

    it('should display correct fields for devices query type with search', () => {
      const query: CatalystVariableQuery = { type: 'devices', search: 'router' };
      render(<VariableQueryEditor query={query} onChange={mockOnChange} />);
      
      const searchInput = screen.getByPlaceholderText(/e.g. branch-a/);
      expect(searchInput).toHaveValue('router');
      
      const definitionInput = screen.getByDisplayValue('devices(search:"router")');
      expect(definitionInput).toBeInTheDocument();
    });

    it('should display correct fields for macs query type', () => {
      const query: CatalystVariableQuery = { type: 'macs' };
      render(<VariableQueryEditor query={query} onChange={mockOnChange} />);
      
      expect(screen.getByText('MACs')).toBeInTheDocument();
      expect(screen.getByText('Search (optional)')).toBeInTheDocument();
      
      const definitionInput = screen.getByDisplayValue('macs()');
      expect(definitionInput).toBeInTheDocument();
    });

    it('should display correct fields for macs query type with search', () => {
      const query: CatalystVariableQuery = { type: 'macs', search: '00:11:22' };
      render(<VariableQueryEditor query={query} onChange={mockOnChange} />);
      
      const searchInput = screen.getByPlaceholderText(/e.g. branch-a/);
      expect(searchInput).toHaveValue('00:11:22');
      
      const definitionInput = screen.getByDisplayValue('macs(search:"00:11:22")');
      expect(definitionInput).toBeInTheDocument();
    });
  });

  describe('Search Input Handling', () => {
    it('should update search field and call onChange for sites', () => {
      const query: CatalystVariableQuery = { type: 'sites' };
      render(<VariableQueryEditor query={query} onChange={mockOnChange} />);
      
      const searchInput = screen.getByPlaceholderText(/e.g. branch-a/);
      fireEvent.change(searchInput, { target: { value: 'branch-a' } });
      
      expect(mockOnChange).toHaveBeenCalledWith(
        { type: 'sites', search: 'branch-a' },
        'sites(search:"branch-a")'
      );
    });

    it('should update search field and call onChange for devices', () => {
      const query: CatalystVariableQuery = { type: 'devices' };
      render(<VariableQueryEditor query={query} onChange={mockOnChange} />);
      
      const searchInput = screen.getByPlaceholderText(/e.g. branch-a/);
      fireEvent.change(searchInput, { target: { value: 'switch-1' } });
      
      expect(mockOnChange).toHaveBeenCalledWith(
        { type: 'devices', search: 'switch-1' },
        'devices(search:"switch-1")'
      );
    });

    it('should update search field and call onChange for macs', () => {
      const query: CatalystVariableQuery = { type: 'macs' };
      render(<VariableQueryEditor query={query} onChange={mockOnChange} />);
      
      const searchInput = screen.getByPlaceholderText(/e.g. branch-a/);
      fireEvent.change(searchInput, { target: { value: 'AA:BB:CC' } });
      
      expect(mockOnChange).toHaveBeenCalledWith(
        { type: 'macs', search: 'AA:BB:CC' },
        'macs(search:"AA:BB:CC")'
      );
    });

    it('should trim whitespace from search input', () => {
      const query: CatalystVariableQuery = { type: 'sites' };
      render(<VariableQueryEditor query={query} onChange={mockOnChange} />);
      
      const searchInput = screen.getByPlaceholderText(/e.g. branch-a/);
      fireEvent.change(searchInput, { target: { value: '  site-name  ' } });
      
      expect(mockOnChange).toHaveBeenCalledWith(
        { type: 'sites', search: 'site-name' },
        'sites(search:"site-name")'
      );
    });

    it('should handle empty search string', () => {
      const query: CatalystVariableQuery = { type: 'sites', search: 'something' };
      render(<VariableQueryEditor query={query} onChange={mockOnChange} />);
      
      const searchInput = screen.getByPlaceholderText(/e.g. branch-a/);
      fireEvent.change(searchInput, { target: { value: '' } });
      
      expect(mockOnChange).toHaveBeenCalledWith(
        { type: 'sites' },
        'sites()'
      );
    });

    it('should handle search string with only whitespace', () => {
      const query: CatalystVariableQuery = { type: 'sites' };
      render(<VariableQueryEditor query={query} onChange={mockOnChange} />);
      
      const searchInput = screen.getByPlaceholderText(/e.g. branch-a/);
      fireEvent.change(searchInput, { target: { value: '   ' } });
      
      expect(mockOnChange).toHaveBeenCalledWith(
        { type: 'sites' },
        'sites()'
      );
    });

    it('should preserve existing search when changing query type', () => {
      const query: CatalystVariableQuery = { type: 'sites', search: 'building' };
      const { rerender } = render(<VariableQueryEditor query={query} onChange={mockOnChange} />);
      
      // Change to devices type - the component should be re-rendered with new query
      const newQuery: CatalystVariableQuery = { type: 'devices', search: 'building' };
      rerender(<VariableQueryEditor query={newQuery} onChange={mockOnChange} />);
      
      const searchInput = screen.getByPlaceholderText(/e.g. branch-a/);
      expect(searchInput).toHaveValue('building');
    });
  });

  describe('Query Definition Display', () => {
    it('should display correct definition for priorities', () => {
      const query: CatalystVariableQuery = { type: 'priorities' };
      render(<VariableQueryEditor query={query} onChange={mockOnChange} />);
      
      const definitionInput = screen.getByDisplayValue('priorities()');
      expect(definitionInput).toHaveAttribute('readonly');
    });

    it('should display correct definition for issueStatuses', () => {
      const query: CatalystVariableQuery = { type: 'issueStatuses' };
      render(<VariableQueryEditor query={query} onChange={mockOnChange} />);
      
      const definitionInput = screen.getByDisplayValue('issueStatuses()');
      expect(definitionInput).toHaveAttribute('readonly');
    });

    it('should display correct definition for sites without search', () => {
      const query: CatalystVariableQuery = { type: 'sites' };
      render(<VariableQueryEditor query={query} onChange={mockOnChange} />);
      
      const definitionInput = screen.getByDisplayValue('sites()');
      expect(definitionInput).toHaveAttribute('readonly');
    });

    it('should display correct definition for sites with search', () => {
      const query: CatalystVariableQuery = { type: 'sites', search: 'test' };
      render(<VariableQueryEditor query={query} onChange={mockOnChange} />);
      
      const definitionInput = screen.getByDisplayValue('sites(search:"test")');
      expect(definitionInput).toHaveAttribute('readonly');
    });

    it('should display correct definition for devices without search', () => {
      const query: CatalystVariableQuery = { type: 'devices' };
      render(<VariableQueryEditor query={query} onChange={mockOnChange} />);
      
      expect(screen.getByDisplayValue('devices()')).toBeInTheDocument();
    });

    it('should display correct definition for devices with search', () => {
      const query: CatalystVariableQuery = { type: 'devices', search: 'router-1' };
      render(<VariableQueryEditor query={query} onChange={mockOnChange} />);
      
      expect(screen.getByDisplayValue('devices(search:"router-1")')).toBeInTheDocument();
    });

    it('should display correct definition for macs without search', () => {
      const query: CatalystVariableQuery = { type: 'macs' };
      render(<VariableQueryEditor query={query} onChange={mockOnChange} />);
      
      expect(screen.getByDisplayValue('macs()')).toBeInTheDocument();
    });

    it('should display correct definition for macs with search', () => {
      const query: CatalystVariableQuery = { type: 'macs', search: 'AA:BB' };
      render(<VariableQueryEditor query={query} onChange={mockOnChange} />);
      
      expect(screen.getByDisplayValue('macs(search:"AA:BB")')).toBeInTheDocument();
    });

    it('should update definition when search changes', () => {
      const query: CatalystVariableQuery = { type: 'sites' };
      const { rerender } = render(<VariableQueryEditor query={query} onChange={mockOnChange} />);
      
      expect(screen.getByDisplayValue('sites()')).toBeInTheDocument();
      
      // Update with search
      const queryWithSearch: CatalystVariableQuery = { type: 'sites', search: 'campus' };
      rerender(<VariableQueryEditor query={queryWithSearch} onChange={mockOnChange} />);
      
      expect(screen.getByDisplayValue('sites(search:"campus")')).toBeInTheDocument();
    });
  });

  describe('Edge Cases', () => {
    it('should handle undefined query gracefully', () => {
      render(<VariableQueryEditor query={undefined} onChange={mockOnChange} />);
      
      // Should default to priorities
      expect(screen.getByText('Priorities (P1..P4)')).toBeInTheDocument();
      expect(screen.getByDisplayValue('priorities()')).toBeInTheDocument();
    });

    it('should handle query without search field gracefully', () => {
      const query: CatalystVariableQuery = { type: 'sites' };
      render(<VariableQueryEditor query={query} onChange={mockOnChange} />);
      
      const searchInput = screen.getByPlaceholderText(/e.g. branch-a/);
      expect(searchInput).toHaveValue('');
    });

    it('should handle switching from type with search to type without search', () => {
      const query: CatalystVariableQuery = { type: 'sites', search: 'test' };
      const { rerender } = render(<VariableQueryEditor query={query} onChange={mockOnChange} />);
      
      expect(screen.getByText('Search (optional)')).toBeInTheDocument();
      
      // Switch to priorities which doesn't have search
      const newQuery: CatalystVariableQuery = { type: 'priorities' };
      rerender(<VariableQueryEditor query={newQuery} onChange={mockOnChange} />);
      
      expect(screen.queryByText('Search (optional)')).not.toBeInTheDocument();
    });

    it('should handle switching from type without search to type with search', () => {
      const query: CatalystVariableQuery = { type: 'priorities' };
      const { rerender } = render(<VariableQueryEditor query={query} onChange={mockOnChange} />);
      
      expect(screen.queryByText('Search (optional)')).not.toBeInTheDocument();
      
      // Switch to sites which has search
      const newQuery: CatalystVariableQuery = { type: 'sites' };
      rerender(<VariableQueryEditor query={newQuery} onChange={mockOnChange} />);
      
      expect(screen.getByText('Search (optional)')).toBeInTheDocument();
    });
  });
});

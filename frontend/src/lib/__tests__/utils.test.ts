import { buildIconUrl } from '../utils';

describe('buildIconUrl', () => {
  const originalEnv = process.env.NEXT_PUBLIC_API_URL;

  afterEach(() => {
    process.env.NEXT_PUBLIC_API_URL = originalEnv;
  });

  it('should return null for null input', () => {
    expect(buildIconUrl(null)).toBeNull();
  });

  it('should return null for undefined input', () => {
    expect(buildIconUrl(undefined)).toBeNull();
  });

  it('should return null for empty string', () => {
    expect(buildIconUrl('')).toBeNull();
  });

  it('should return null for whitespace-only string', () => {
    expect(buildIconUrl(' ')).toBeNull();
    expect(buildIconUrl('  ')).toBeNull();
    expect(buildIconUrl('\t')).toBeNull();
    expect(buildIconUrl('\n')).toBeNull();
  });

  it('should build URL with default API URL', () => {
    delete process.env.NEXT_PUBLIC_API_URL;
    expect(buildIconUrl('/uploads/icons/user123.png')).toBe(
      'http://localhost:8080/uploads/icons/user123.png'
    );
  });

  it('should build URL with custom API URL', () => {
    process.env.NEXT_PUBLIC_API_URL = 'https://api.example.com';
    expect(buildIconUrl('/uploads/icons/user123.png')).toBe(
      'https://api.example.com/uploads/icons/user123.png'
    );
  });

  it('should handle paths without leading slash', () => {
    delete process.env.NEXT_PUBLIC_API_URL;
    expect(buildIconUrl('uploads/icons/user123.png')).toBe(
      'http://localhost:8080uploads/icons/user123.png'
    );
  });

  it('should preserve path with spaces (after trim check)', () => {
    delete process.env.NEXT_PUBLIC_API_URL;
    expect(buildIconUrl('/uploads/icons/user 123.png')).toBe(
      'http://localhost:8080/uploads/icons/user 123.png'
    );
  });
});

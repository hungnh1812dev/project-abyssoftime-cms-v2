const levels: Record<string, number> = {
  super_admin: 4,
  admin: 3,
  editor: 2,
  guest: 1,
};

export function roleLevel(role: string | null): number {
  if (!role) return 0;
  return levels[role] ?? 0;
}

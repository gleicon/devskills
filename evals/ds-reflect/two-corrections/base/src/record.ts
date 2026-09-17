export type Record = {
  id: string;
  owner_id: string;
  created_at: number;
};

export function age(r: Record, now: number): number {
  return now - r.created_at;
}

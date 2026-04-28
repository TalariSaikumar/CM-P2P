export type User = {
  id: string;
  email: string;
  role: "CUSTOMER" | "OWNER";
  full_name: string;
  phone_number: string;
  address: string;
  aadhaar_number: string;
  is_kyc_verified: boolean;
  driving_license_number?: string | null;
};

const TOKEN_KEY = "carmanage_token";
const USER_KEY = "carmanage_user";

export function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(TOKEN_KEY);
}

export function getUser(): User | null {
  if (typeof window === "undefined") return null;
  const raw = localStorage.getItem(USER_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as User;
  } catch {
    return null;
  }
}

export function setSession(token: string, user: User) {
  localStorage.setItem(TOKEN_KEY, token);
  localStorage.setItem(USER_KEY, JSON.stringify(user));
}

export function clearSession() {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(USER_KEY);
}

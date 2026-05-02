"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { apiJson, ApiError } from "@/lib/api";
import { setSession, type User } from "@/lib/session";
import { ButtonCarSpinner } from "@/components/loaders";

type RegisterResponse = { token: string; user: User };

export default function RegisterPage() {
  const router = useRouter();
  const [form, setForm] = useState({
    email: "",
    password: "",
    role: "CUSTOMER" as "CUSTOMER" | "OWNER",
    full_name: "",
    aadhaar_number: "",
    phone_number: "",
    address: "",
    driving_license_number: "",
  });
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      const body: Record<string, unknown> = {
        email: form.email,
        password: form.password,
        role: form.role,
        full_name: form.full_name,
        aadhaar_number: form.aadhaar_number,
        phone_number: form.phone_number,
        address: form.address,
      };
      if (form.driving_license_number.trim()) {
        body.driving_license_number = form.driving_license_number.trim();
      }
      const res = await apiJson<RegisterResponse>("/auth/register", {
        method: "POST",
        body: JSON.stringify(body),
        token: null,
      });
      setSession(res.token, res.user);
      router.push(res.user.role === "OWNER" ? "/owner/fleet" : "/customer/search");
      router.refresh();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Registration failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="page-shell max-w-lg">
      <h1 className="text-xl font-semibold sm:text-2xl">Create account</h1>
      <p className="mt-1 text-sm text-slate-600">
        Already registered?{" "}
        <Link href="/login" className="text-slate-900 underline">
          Sign in
        </Link>
      </p>
      <form onSubmit={onSubmit} className="mt-6 space-y-4">
        {error && (
          <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
            {error}
          </div>
        )}
        <div>
          <label className="block text-sm font-medium text-slate-700">I am a</label>
          <select
            className="mt-1 w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
            value={form.role}
            onChange={(e) =>
              setForm((f) => ({ ...f, role: e.target.value as "CUSTOMER" | "OWNER" }))
            }
          >
            <option value="CUSTOMER">Customer (book cars)</option>
            <option value="OWNER">Owner (list cars)</option>
          </select>
        </div>
        <div>
          <label className="block text-sm font-medium text-slate-700">Email</label>
          <input
            className="mt-1 w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
            type="email"
            value={form.email}
            onChange={(e) => setForm((f) => ({ ...f, email: e.target.value }))}
            required
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-slate-700">Password (min 8)</label>
          <input
            className="mt-1 w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
            type="password"
            value={form.password}
            onChange={(e) => setForm((f) => ({ ...f, password: e.target.value }))}
            minLength={8}
            required
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-slate-700">Full name</label>
          <input
            className="mt-1 w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
            value={form.full_name}
            onChange={(e) => setForm((f) => ({ ...f, full_name: e.target.value }))}
            required
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-slate-700">Aadhaar number</label>
          <input
            className="mt-1 w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
            value={form.aadhaar_number}
            onChange={(e) => setForm((f) => ({ ...f, aadhaar_number: e.target.value }))}
            required
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-slate-700">Phone number</label>
          <input
            className="mt-1 w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
            value={form.phone_number}
            onChange={(e) => setForm((f) => ({ ...f, phone_number: e.target.value }))}
            required
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-slate-700">Address</label>
          <textarea
            className="mt-1 w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
            rows={3}
            value={form.address}
            onChange={(e) => setForm((f) => ({ ...f, address: e.target.value }))}
            required
          />
        </div>
        {form.role === "CUSTOMER" && (
          <div>
            <label className="block text-sm font-medium text-slate-700">
              Driving license number (required before booking)
            </label>
            <input
              className="mt-1 w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
              value={form.driving_license_number}
              onChange={(e) =>
                setForm((f) => ({ ...f, driving_license_number: e.target.value }))
              }
            />
          </div>
        )}
        <button
          type="submit"
          disabled={loading}
          className="inline-flex w-full items-center justify-center gap-2 rounded-md bg-slate-900 py-2 text-sm font-medium text-white hover:bg-slate-800 disabled:opacity-60"
        >
          {loading ? (
            <>
              <ButtonCarSpinner className="text-emerald-200" />
              Creating account…
            </>
          ) : (
            "Create account"
          )}
        </button>
      </form>
    </main>
  );
}

"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { apiJson, ApiError } from "@/lib/api";
import { getToken, getUser, setSession, type User } from "@/lib/session";

type MeResponse = { user: User };

export default function AccountPage() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [msg, setMsg] = useState<string | null>(null);
  const [err, setErr] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!getToken()) {
      router.replace("/login");
      return;
    }
    setUser(getUser());
    void refreshMe();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  async function refreshMe() {
    try {
      const res = await apiJson<MeResponse>("/me");
      setUser(res.user);
      const t = getToken();
      if (t) setSession(t, res.user);
    } catch {
      /* ignore */
    }
  }

  async function completeKyc() {
    setErr(null);
    setMsg(null);
    setLoading(true);
    try {
      const res = await apiJson<MeResponse>("/me/complete-kyc", { method: "POST" });
      setUser(res.user);
      const t = getToken();
      if (t) setSession(t, res.user);
      setMsg("Your profile is now marked verified on this environment.");
    } catch (e) {
      setErr(e instanceof ApiError ? e.message : "Could not complete verification");
    } finally {
      setLoading(false);
    }
  }

  async function saveProfile(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setErr(null);
    setMsg(null);
    const fd = new FormData(e.currentTarget);
    const body: Record<string, unknown> = {
      full_name: String(fd.get("full_name") || ""),
      phone_number: String(fd.get("phone_number") || ""),
      address: String(fd.get("address") || ""),
    };
    if (user?.role === "CUSTOMER") {
      const dl = String(fd.get("driving_license_number") || "").trim();
      body.driving_license_number = dl.length ? dl : null;
    }
    setLoading(true);
    try {
      const res = await apiJson<MeResponse>("/me", {
        method: "PUT",
        body: JSON.stringify(body),
      });
      setUser(res.user);
      const t = getToken();
      if (t) setSession(t, res.user);
      setMsg("Profile updated.");
    } catch (e) {
      setErr(e instanceof ApiError ? e.message : "Update failed");
    } finally {
      setLoading(false);
    }
  }

  if (!user) {
    return (
      <main className="page-shell max-w-lg">
        <p className="text-slate-600">Loading…</p>
      </main>
    );
  }

  return (
    <main className="page-shell max-w-lg space-y-6">
      <div>
        <h1 className="text-xl font-semibold sm:text-2xl">Account</h1>
        <p className="text-sm text-slate-600">
          KYC status:{" "}
          <span className="font-medium text-slate-900">
            {user.is_kyc_verified ? "Verified" : "Not verified"}
          </span>
        </p>
      </div>
      {msg && (
        <div className="rounded-md border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm text-emerald-900">
          {msg}
        </div>
      )}
      {err && (
        <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
          {err}
        </div>
      )}
      <form onSubmit={saveProfile} className="space-y-3 rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
        <h2 className="font-medium">Profile</h2>
        <div>
          <label className="text-sm text-slate-700">Full name</label>
          <input
            name="full_name"
            defaultValue={user.full_name}
            className="mt-1 w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
          />
        </div>
        <div>
          <label className="text-sm text-slate-700">Phone</label>
          <input
            name="phone_number"
            defaultValue={user.phone_number}
            className="mt-1 w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
          />
        </div>
        <div>
          <label className="text-sm text-slate-700">Address</label>
          <textarea
            name="address"
            rows={3}
            className="mt-1 w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
            defaultValue={user.address}
          />
        </div>
        {user.role === "CUSTOMER" && (
          <div>
            <label className="text-sm text-slate-700">Driving license number</label>
            <input
              name="driving_license_number"
              defaultValue={user.driving_license_number || ""}
              className="mt-1 w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
            />
          </div>
        )}
        <button
          type="submit"
          disabled={loading}
          className="rounded-md bg-slate-900 px-4 py-2 text-sm text-white hover:bg-slate-800 disabled:opacity-60"
        >
          Save changes
        </button>
      </form>

      {!user.is_kyc_verified && (
        <div className="rounded-lg border border-amber-200 bg-amber-50 p-4 text-sm text-amber-950">
          <p className="font-medium">Demo verification</p>
          <p className="mt-1">
            When the API has <code className="rounded bg-amber-100 px-1">ALLOW_SELF_KYC_VERIFY=true</code>, you
            can mark yourself verified for local testing.
          </p>
          <button
            type="button"
            disabled={loading}
            onClick={completeKyc}
            className="mt-3 rounded-md bg-amber-900 px-3 py-1.5 text-white hover:bg-amber-800 disabled:opacity-60"
          >
            Mark KYC verified (demo)
          </button>
        </div>
      )}
    </main>
  );
}

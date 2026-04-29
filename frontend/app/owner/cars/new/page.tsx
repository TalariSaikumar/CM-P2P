"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { apiJson, ApiError } from "@/lib/api";
import { getToken } from "@/lib/session";

export default function NewCarPage() {
  const router = useRouter();
  const [car_name, setCarName] = useState("");
  const [car_model, setCarModel] = useState("");
  const [car_number, setCarNumber] = useState("");
  const [registration_number, setReg] = useState("");
  const [engine_number, setEngine] = useState("");
  const [price_per_hour, setPH] = useState("100");
  const [price_per_day, setPD] = useState("1500");
  const [price_per_km, setPK] = useState("15");
  const [location, setLocation] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!getToken()) {
      router.push("/login");
      return;
    }
    setError(null);
    setLoading(true);
    try {
      await apiJson("/cars", {
        method: "POST",
        body: JSON.stringify({
          car_name,
          car_model,
          car_number,
          registration_number,
          engine_number,
          price_per_hour,
          price_per_day,
          price_per_km,
          location,
        }),
      });
      router.push("/owner/fleet");
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not create listing");
    } finally {
      setLoading(false);
    }
  }

  const field =
    "mt-1 w-full rounded-md border border-slate-300 px-3 py-2 text-sm focus:border-slate-500 focus:outline-none focus:ring-1 focus:ring-slate-500";

  return (
    <main className="page-shell max-w-lg">
      <h1 className="text-xl font-semibold sm:text-2xl">List a vehicle</h1>
      <p className="mt-1 text-sm text-slate-600">Owners must be KYC verified before listings are accepted.</p>
      <form onSubmit={onSubmit} className="mt-6 space-y-3">
        {error && (
          <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">{error}</div>
        )}
        <div>
          <label className="text-sm text-slate-700">Car name</label>
          <input className={field} value={car_name} onChange={(e) => setCarName(e.target.value)} required />
        </div>
        <div>
          <label className="text-sm text-slate-700">Model</label>
          <input className={field} value={car_model} onChange={(e) => setCarModel(e.target.value)} required />
        </div>
        <div>
          <label className="text-sm text-slate-700">Plate number</label>
          <input className={field} value={car_number} onChange={(e) => setCarNumber(e.target.value)} required />
        </div>
        <div>
          <label className="text-sm text-slate-700">RC number</label>
          <input className={field} value={registration_number} onChange={(e) => setReg(e.target.value)} required />
        </div>
        <div>
          <label className="text-sm text-slate-700">Engine number</label>
          <input className={field} value={engine_number} onChange={(e) => setEngine(e.target.value)} required />
        </div>
        <div>
          <label className="text-sm text-slate-700">Location (city/area)</label>
          <input className={field} value={location} onChange={(e) => setLocation(e.target.value)} required />
        </div>
        <div>
          <label className="text-sm text-slate-700">Price per hour (INR)</label>
          <input className={field} value={price_per_hour} onChange={(e) => setPH(e.target.value)} required />
        </div>
        <div>
          <label className="text-sm text-slate-700">Price per day (INR)</label>
          <input className={field} value={price_per_day} onChange={(e) => setPD(e.target.value)} required />
        </div>
        <div>
          <label className="text-sm text-slate-700">Price per km (INR)</label>
          <input className={field} value={price_per_km} onChange={(e) => setPK(e.target.value)} required />
        </div>
        <button
          type="submit"
          disabled={loading}
          className="w-full rounded-md bg-slate-900 py-2 text-sm text-white hover:bg-slate-800 disabled:opacity-60"
        >
          {loading ? "Saving…" : "Create listing"}
        </button>
      </form>
    </main>
  );
}

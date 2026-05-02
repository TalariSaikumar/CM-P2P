"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { apiJson, ApiError } from "@/lib/api";
import { getToken } from "@/lib/session";
import type { CarMineRow } from "@/lib/apitypes";
import { ButtonCarSpinner, PageLoader } from "@/components/loaders";

type AirbagRow = { type: string; count: string };

const currentYear = new Date().getFullYear();

export default function EditCarPage() {
  const router = useRouter();
  const params = useParams<{ id: string }>();
  const carId = params?.id ?? "";

  const [loadError, setLoadError] = useState<string | null>(null);
  const [loadingCar, setLoadingCar] = useState(true);
  const [bookedToday, setBookedToday] = useState(false);

  const [car_name, setCarName] = useState("");
  const [car_model, setCarModel] = useState("");
  const [car_number, setCarNumber] = useState("");
  const [registration_number, setReg] = useState("");
  const [engine_number, setEngine] = useState("");
  const [price_per_hour, setPH] = useState("");
  const [price_per_day, setPD] = useState("");
  const [price_per_km, setPK] = useState("");
  const [location, setLocation] = useState("");
  const [is_active, setIsActive] = useState(true);

  const [model_year, setModelYear] = useState(String(currentYear));
  const [color, setColor] = useState("");
  const [fuel_type, setFuelType] = useState("petrol");
  const [transmission, setTransmission] = useState("manual");
  const [mileage_km, setMileageKm] = useState("");
  const [num_seats, setNumSeats] = useState("5");
  const [airbags, setAirbags] = useState(false);
  const [airbagRows, setAirbagRows] = useState<AirbagRow[]>([{ type: "Front", count: "2" }]);
  const [camera_type, setCameraType] = useState("");
  const [air_conditioning, setAirConditioning] = useState(false);
  const [cruise_control, setCruiseControl] = useState(false);
  const [open_roof, setOpenRoof] = useState(false);
  const [navigation, setNavigation] = useState(false);
  const [speakers, setSpeakers] = useState(false);

  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  const hydrate = useCallback((data: CarMineRow) => {
    setBookedToday(data.booked_for_current_date);
    setCarName(data.car_name);
    setCarModel(data.car_model);
    setCarNumber(data.car_number);
    setReg(data.registration_number);
    setEngine(data.engine_number);
    setPH(data.price_per_hour);
    setPD(data.price_per_day);
    setPK(data.price_per_km);
    setLocation(data.location);
    setIsActive(data.is_active);
    setModelYear(String(data.model_year));
    setColor(data.color);
    setFuelType(data.fuel_type);
    setTransmission(data.transmission);
    setMileageKm(String(data.mileage_km));
    setNumSeats(String(data.num_seats));
    setAirbags(data.airbags);
    if (data.airbags && data.airbag_details?.length) {
      setAirbagRows(data.airbag_details.map((d) => ({ type: d.type, count: String(d.count) })));
    } else {
      setAirbagRows([{ type: "Front", count: "2" }]);
    }
    setCameraType(data.camera_type ?? "");
    setAirConditioning(data.air_conditioning);
    setCruiseControl(data.cruise_control);
    setOpenRoof(data.open_roof);
    setNavigation(data.navigation);
    setSpeakers(data.speakers);
  }, []);

  useEffect(() => {
    if (!getToken()) {
      router.replace("/login");
      return;
    }
    if (!carId) {
      setLoadError("Missing vehicle id.");
      setLoadingCar(false);
      return;
    }
    let cancelled = false;
    (async () => {
      setLoadError(null);
      setLoadingCar(true);
      try {
        const data = await apiJson<CarMineRow>(`/cars/${carId}/edit`);
        if (cancelled) return;
        hydrate(data);
      } catch (e) {
        if (!cancelled) {
          setLoadError(e instanceof ApiError ? e.message : "Could not load vehicle");
        }
      } finally {
        if (!cancelled) setLoadingCar(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [carId, router, hydrate]);

  function addAirbagRow() {
    setAirbagRows((rows) => [...rows, { type: "", count: "1" }]);
  }

  function removeAirbagRow(i: number) {
    setAirbagRows((rows) => rows.filter((_, idx) => idx !== i));
  }

  function updateAirbagRow(i: number, patch: Partial<AirbagRow>) {
    setAirbagRows((rows) => rows.map((r, idx) => (idx === i ? { ...r, ...patch } : r)));
  }

  const formLocked = bookedToday;

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!getToken()) {
      router.push("/login");
      return;
    }
    if (formLocked) return;
    setError(null);

    const my = parseInt(model_year, 10);
    if (Number.isNaN(my) || my < 1980 || my > currentYear + 2) {
      setError("Enter a valid model year.");
      return;
    }
    const seats = parseInt(num_seats, 10);
    if (Number.isNaN(seats) || seats < 1 || seats > 20) {
      setError("Number of seats must be between 1 and 20.");
      return;
    }
    const mileage = parseInt(mileage_km, 10);
    if (Number.isNaN(mileage) || mileage < 0) {
      setError("Enter mileage in km (0 or greater).");
      return;
    }

    let airbag_details: { type: string; count: number }[] = [];
    let airbag_count = 0;
    if (airbags) {
      airbag_details = airbagRows
        .map((r) => ({
          type: r.type.trim(),
          count: parseInt(r.count, 10) || 0,
        }))
        .filter((r) => r.type.length > 0);
      if (!airbag_details.length) {
        setError("With airbags enabled, add at least one airbag type and count.");
        return;
      }
      if (airbag_details.some((r) => r.count < 1)) {
        setError("Each airbag line needs a count of at least 1.");
        return;
      }
      airbag_count = airbag_details.reduce((s, r) => s + r.count, 0);
    }

    setSaving(true);
    try {
      await apiJson(`/cars/${carId}`, {
        method: "PATCH",
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
          is_active,
          model_year: my,
          color,
          fuel_type,
          transmission,
          mileage_km: mileage,
          num_seats: seats,
          airbags,
          airbag_count,
          airbag_details,
          camera_type,
          air_conditioning,
          cruise_control,
          open_roof,
          navigation,
          speakers,
        }),
      });
      router.push("/owner/fleet");
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not save changes");
    } finally {
      setSaving(false);
    }
  }

  const field =
    "mt-1 w-full rounded-md border border-slate-300 px-3 py-2 text-sm focus:border-slate-500 focus:outline-none focus:ring-1 focus:ring-slate-500 disabled:bg-slate-100 disabled:text-slate-500";
  const check =
    "h-4 w-4 rounded border-slate-300 text-slate-900 focus:ring-slate-500 disabled:opacity-50";

  if (loadingCar) {
    return (
      <main className="page-shell max-w-2xl">
        <PageLoader title="Loading vehicle…" subtitle="Specs, pricing, and availability." className="min-h-[260px] py-8" />
      </main>
    );
  }

  if (loadError) {
    return (
      <main className="page-shell max-w-2xl space-y-4">
        <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">{loadError}</div>
        <Link href="/owner/fleet" className="text-sm font-medium text-emerald-800 hover:text-emerald-900">
          ← Back to fleet
        </Link>
      </main>
    );
  }

  return (
    <main className="page-shell max-w-2xl">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h1 className="text-xl font-semibold sm:text-2xl">Edit vehicle</h1>
        <Link href="/owner/fleet" className="text-sm font-medium text-slate-600 hover:text-slate-900">
          ← Fleet
        </Link>
      </div>
      <p className="mt-1 text-sm text-slate-600">Only you can change this listing. Updates are blocked when the car is booked for the current UTC calendar day.</p>

      {formLocked && (
        <div className="mt-4 rounded-md border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-950">
          This vehicle has an active booking that includes today (UTC). Saving changes is disabled until that rental
          window no longer covers today.
        </div>
      )}

      <form onSubmit={onSubmit} className="mt-6 space-y-4">
        {error && (
          <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">{error}</div>
        )}

        <fieldset className="space-y-3 rounded-lg border border-slate-200 p-4" disabled={formLocked}>
          <legend className="px-1 text-sm font-medium text-slate-800">Identity and pricing</legend>
          <label className="flex items-center gap-2 text-sm text-slate-800">
            <input type="checkbox" className={check} checked={is_active} onChange={(e) => setIsActive(e.target.checked)} />
            Listing active (visible in search)
          </label>
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
          <div className="grid gap-3 sm:grid-cols-3">
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
          </div>
        </fieldset>

        <fieldset className="space-y-3 rounded-lg border border-slate-200 p-4" disabled={formLocked}>
          <legend className="px-1 text-sm font-medium text-slate-800">Vehicle details</legend>
          <div className="grid gap-3 sm:grid-cols-2">
            <div>
              <label className="text-sm text-slate-700">Model year</label>
              <input
                className={field}
                type="number"
                min={1980}
                max={currentYear + 2}
                value={model_year}
                onChange={(e) => setModelYear(e.target.value)}
                required
              />
            </div>
            <div>
              <label className="text-sm text-slate-700">Color</label>
              <input className={field} value={color} onChange={(e) => setColor(e.target.value)} required />
            </div>
            <div>
              <label className="text-sm text-slate-700">Fuel type</label>
              <select className={field} value={fuel_type} onChange={(e) => setFuelType(e.target.value)} required>
                <option value="petrol">Petrol</option>
                <option value="diesel">Diesel</option>
                <option value="cng">CNG</option>
                <option value="ev">EV</option>
              </select>
            </div>
            <div>
              <label className="text-sm text-slate-700">Transmission</label>
              <select
                className={field}
                value={transmission}
                onChange={(e) => setTransmission(e.target.value)}
                required
              >
                <option value="manual">Manual</option>
                <option value="auto">Automatic</option>
              </select>
            </div>
            <div>
              <label className="text-sm text-slate-700">Mileage (km)</label>
              <input
                className={field}
                type="number"
                min={0}
                value={mileage_km}
                onChange={(e) => setMileageKm(e.target.value)}
                required
              />
            </div>
            <div>
              <label className="text-sm text-slate-700">Number of seats</label>
              <input
                className={field}
                type="number"
                min={1}
                max={20}
                value={num_seats}
                onChange={(e) => setNumSeats(e.target.value)}
                required
              />
            </div>
            <div className="sm:col-span-2">
              <label className="text-sm text-slate-700">Camera type</label>
              <input
                className={field}
                value={camera_type}
                onChange={(e) => setCameraType(e.target.value)}
                placeholder="e.g. none, rear, front and rear, 360°"
              />
            </div>
          </div>

          <div className="rounded-md border border-slate-100 bg-slate-50/80 p-3">
            <label className="flex items-center gap-2 text-sm text-slate-800">
              <input type="checkbox" className={check} checked={airbags} onChange={(e) => setAirbags(e.target.checked)} />
              Airbags
            </label>
            {airbags && (
              <div className="mt-3 space-y-2">
                <p className="text-xs text-slate-600">
                  List each airbag type and how many. Counts must add up to the total number of airbags.
                </p>
                {airbagRows.map((row, i) => (
                  <div key={i} className="flex flex-wrap items-end gap-2">
                    <div className="min-w-0 flex-1">
                      <label className="text-xs text-slate-600">Type</label>
                      <input
                        className={field}
                        value={row.type}
                        onChange={(e) => updateAirbagRow(i, { type: e.target.value })}
                        placeholder="e.g. Side curtain"
                      />
                    </div>
                    <div className="w-24 shrink-0">
                      <label className="text-xs text-slate-600">Count</label>
                      <input
                        className={field}
                        type="number"
                        min={1}
                        value={row.count}
                        onChange={(e) => updateAirbagRow(i, { count: e.target.value })}
                      />
                    </div>
                    <button
                      type="button"
                      className="mb-0.5 rounded border border-slate-300 px-2 py-1 text-xs text-slate-700 hover:bg-white"
                      onClick={() => removeAirbagRow(i)}
                      disabled={airbagRows.length <= 1}
                    >
                      Remove
                    </button>
                  </div>
                ))}
                <button
                  type="button"
                  className="text-sm font-medium text-emerald-800 hover:text-emerald-900"
                  onClick={addAirbagRow}
                >
                  + Add airbag type
                </button>
              </div>
            )}
          </div>

          <div>
            <p className="text-sm font-medium text-slate-800">Features</p>
            <div className="mt-2 grid gap-2 sm:grid-cols-2">
              <label className="flex items-center gap-2 text-sm text-slate-700">
                <input
                  type="checkbox"
                  className={check}
                  checked={air_conditioning}
                  onChange={(e) => setAirConditioning(e.target.checked)}
                />
                Air conditioning
              </label>
              <label className="flex items-center gap-2 text-sm text-slate-700">
                <input
                  type="checkbox"
                  className={check}
                  checked={cruise_control}
                  onChange={(e) => setCruiseControl(e.target.checked)}
                />
                Cruise control
              </label>
              <label className="flex items-center gap-2 text-sm text-slate-700">
                <input type="checkbox" className={check} checked={open_roof} onChange={(e) => setOpenRoof(e.target.checked)} />
                Open roof (sunroof)
              </label>
              <label className="flex items-center gap-2 text-sm text-slate-700">
                <input
                  type="checkbox"
                  className={check}
                  checked={navigation}
                  onChange={(e) => setNavigation(e.target.checked)}
                />
                Navigation
              </label>
              <label className="flex items-center gap-2 text-sm text-slate-700">
                <input type="checkbox" className={check} checked={speakers} onChange={(e) => setSpeakers(e.target.checked)} />
                Speakers
              </label>
            </div>
          </div>
        </fieldset>

        <button
          type="submit"
          disabled={formLocked || saving}
          className="inline-flex w-full items-center justify-center gap-2 rounded-md bg-slate-900 py-2 text-sm text-white hover:bg-slate-800 disabled:opacity-60"
        >
          {saving ? (
            <>
              <ButtonCarSpinner className="text-emerald-200" />
              Saving…
            </>
          ) : (
            "Save changes"
          )}
        </button>
      </form>
    </main>
  );
}

"use client";

import { useEffect, useState, useCallback } from "react";
import { useRouter } from "next/navigation";
import { apiJson, ApiError } from "@/lib/api";
import { getToken, getUser } from "@/lib/session";
import type { Car } from "@/lib/apitypes";
import type { User } from "@/lib/session";
import { PaginationBar } from "@/components/PaginationBar";
import { ButtonCarSpinner, SearchResultsSkeleton } from "@/components/loaders";

function isOwnCar(user: User | null, car: Car): boolean {
  return !!user && user.id === car.owner_id;
}

function toStartOfDayUTC(dateStr: string): string {
  return `${dateStr}T00:00:00.000Z`;
}

function toEndOfDayUTC(dateStr: string): string {
  return `${dateStr}T23:59:59.999Z`;
}

const PER_PAGE = 20;

export default function CustomerSearchPage() {
  const router = useRouter();
  const [location, setLocation] = useState("");
  const [model, setModel] = useState("");
  const [cars, setCars] = useState<Car[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  /** Bumped when the user clicks Search so we refetch even if already on page 1. */
  const [searchNonce, setSearchNonce] = useState(0);
  const [error, setError] = useState<string | null>(null);
  const [searchLoading, setSearchLoading] = useState(true);
  const [submitLoading, setSubmitLoading] = useState(false);

  const [bookingCar, setBookingCar] = useState<Car | null>(null);
  const [rentalFrom, setRentalFrom] = useState("");
  const [rentalTo, setRentalTo] = useState("");
  const [pickupPoint, setPickupPoint] = useState("");
  const [dropPoint, setDropPoint] = useState("");
  const [customerNote, setCustomerNote] = useState("");

  const loadPage = useCallback(
    async (p: number) => {
      setError(null);
      setSearchLoading(true);
      try {
        const qs = new URLSearchParams();
        if (location.trim()) qs.set("location", location.trim());
        if (model.trim()) qs.set("model", model.trim());
        qs.set("page", String(p));
        qs.set("per_page", String(PER_PAGE));
        const res = await apiJson<{ cars: Car[]; total?: number }>(`/cars?${qs.toString()}`);
        const t = res.total ?? 0;
        setTotal(t);
        const totalPages = Math.max(1, Math.ceil(t / PER_PAGE));
        if (p > totalPages) {
          setPage(totalPages);
          return;
        }
        setCars(res.cars || []);
      } catch (e) {
        setError(e instanceof ApiError ? e.message : "Search failed");
      } finally {
        setSearchLoading(false);
      }
    },
    [location, model],
  );

  useEffect(() => {
    void loadPage(page);
  }, [page, searchNonce, loadPage]);

  function openBookingModal(car: Car) {
    const token = getToken();
    const user = getUser();
    if (!token || !user) {
      router.push("/login");
      return;
    }
    if (isOwnCar(user, car)) {
      setError("You cannot book your own car.");
      return;
    }
    if (user.role !== "CUSTOMER") {
      setError("Only customers can create bookings.");
      return;
    }
    if (!user.is_kyc_verified) {
      setError("Complete KYC on the Account page before booking.");
      return;
    }
    if (!user.driving_license_number) {
      setError("Add your driving license number on the Account page before booking.");
      return;
    }
    setError(null);
    setBookingCar(car);
    const today = new Date().toISOString().slice(0, 10);
    setRentalFrom(today);
    setRentalTo(today);
    setPickupPoint("");
    setDropPoint("");
    setCustomerNote("");
  }

  function closeBookingModal() {
    setBookingCar(null);
    setError(null);
  }

  async function submitBooking() {
    if (!bookingCar) return;
    if (!rentalFrom || !rentalTo) {
      setError("Please choose rental start and end dates.");
      return;
    }
    if (rentalTo < rentalFrom) {
      setError("End date must be on or after the start date.");
      return;
    }
    setError(null);
    setSubmitLoading(true);
    try {
      const res = await apiJson<{ booking: { id: string } }>("/bookings", {
        method: "POST",
        body: JSON.stringify({
          car_id: bookingCar.id,
          customer_note: customerNote.trim(),
          rental_from: toStartOfDayUTC(rentalFrom),
          rental_to: toEndOfDayUTC(rentalTo),
          pickup_point: pickupPoint.trim(),
          drop_point: dropPoint.trim(),
        }),
      });
      closeBookingModal();
      router.push(`/bookings/${res.booking.id}`);
    } catch (e) {
      if (e instanceof ApiError && e.code === "CAR_ALREADY_BOOKED") {
        setError("Already booked for those dates. Pick other dates or another car.");
      } else {
        setError(e instanceof ApiError ? e.message : "Could not create booking");
      }
    } finally {
      setSubmitLoading(false);
    }
  }

  return (
    <main className="page-shell max-w-4xl space-y-6">
      <div>
        <h1 className="text-xl font-semibold sm:text-2xl">Search cars</h1>
        <p className="text-sm text-slate-600">
          Filter by location and model, then start a booking. Rental dates are required; pickup and drop-off can be
          added later on the booking page. The owner confirms after you agree on price (chat is optional).
        </p>
      </div>
      <div className="flex flex-col gap-3 rounded-lg border border-slate-200 bg-white p-4 shadow-sm sm:flex-row sm:flex-wrap sm:items-end">
        <input
          placeholder="Location (city or area)"
          className="min-h-[44px] min-w-0 flex-1 rounded-md border border-slate-300 px-3 py-2 text-sm sm:min-w-[140px]"
          value={location}
          onChange={(e) => setLocation(e.target.value)}
        />
        <input
          placeholder="Model keyword"
          className="min-h-[44px] min-w-0 flex-1 rounded-md border border-slate-300 px-3 py-2 text-sm sm:min-w-[120px]"
          value={model}
          onChange={(e) => setModel(e.target.value)}
        />
        <button
          type="button"
          onClick={() => {
            setPage(1);
            setSearchNonce((n) => n + 1);
          }}
          disabled={searchLoading}
          className="inline-flex min-h-[44px] w-full shrink-0 items-center justify-center gap-2 rounded-md bg-slate-900 px-4 py-2 text-sm text-white hover:bg-slate-800 disabled:opacity-60 sm:w-auto"
        >
          {searchLoading ? (
            <>
              <ButtonCarSpinner className="text-emerald-200" />
              Searching…
            </>
          ) : (
            "Search"
          )}
        </button>
      </div>
      {error && !bookingCar && (
        <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">{error}</div>
      )}
      {searchLoading ? (
        <SearchResultsSkeleton rows={5} />
      ) : (
      <ul className="space-y-3">
        {cars.map((c) => {
          const own = isOwnCar(getUser(), c);
          return (
            <li
              key={c.id}
              className="flex flex-col gap-3 rounded-lg border border-slate-200 bg-white p-4 shadow-sm sm:flex-row sm:items-center sm:justify-between"
            >
              <div>
                <p className="font-medium text-slate-900">
                  {c.car_name} · {c.car_model}
                </p>
                <p className="text-sm text-slate-600">
                  {c.model_year} · {c.color} · {c.fuel_type} · {c.transmission} · {c.num_seats} seats
                </p>
                <p className="text-sm text-slate-600">
                  {c.location} · Plate {c.car_number}
                </p>
                <p className="text-sm text-slate-600">
                  From ₹{c.price_per_day}/day · ₹{c.price_per_hour}/hr · ₹{c.price_per_km}/km
                </p>
              </div>
              <div className="flex w-full shrink-0 sm:w-auto">
                {own ? (
                  <span
                    className="inline-flex min-h-[44px] w-full items-center justify-center rounded-md border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-500 sm:min-h-0 sm:w-auto sm:py-1.5"
                    title="This listing belongs to your account"
                  >
                    Your listing
                  </span>
                ) : (
                  <button
                    type="button"
                    onClick={() => openBookingModal(c)}
                    className="min-h-[44px] w-full rounded-md border border-slate-300 px-3 py-2 text-sm hover:bg-slate-50 sm:min-h-0 sm:w-auto sm:py-1.5"
                  >
                    Book
                  </button>
                )}
              </div>
            </li>
          );
        })}
        {!cars.length && (
          <p className="text-sm text-slate-600">No cars match your filters yet.</p>
        )}
      </ul>
      )}
      <PaginationBar page={page} perPage={PER_PAGE} total={total} onPageChange={setPage} noun="cars" />

      {bookingCar && (
        <div
          className="fixed inset-0 z-50 flex items-end justify-center bg-black/40 p-4 sm:items-center"
          onClick={closeBookingModal}
          role="presentation"
        >
          <div
            className="max-h-[90dvh] w-full max-w-lg overflow-y-auto rounded-xl border border-slate-200 bg-white p-5 shadow-xl"
            role="dialog"
            aria-labelledby="booking-dialog-title"
            onClick={(e) => e.stopPropagation()}
          >
            <h2 id="booking-dialog-title" className="text-lg font-semibold text-slate-900">
              Book {bookingCar.car_name} · {bookingCar.car_model}
            </h2>
            <p className="mt-1 text-sm text-slate-600">
              Rental dates are required. Pickup and drop-off are optional here—you can set them on the booking page
              before the trip is finalized. You can chat to negotiate price; only the owner can confirm the booking.
            </p>
            {error && (
              <div
                className="mt-4 rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800"
                role="alert"
              >
                {error}
              </div>
            )}
            <div className="mt-4 space-y-3">
              <div className="grid gap-3 sm:grid-cols-2">
                <div>
                  <label className="text-sm text-slate-700">From date</label>
                  <input
                    type="date"
                    className="mt-1 w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                    value={rentalFrom}
                    onChange={(e) => {
                      setError(null);
                      setRentalFrom(e.target.value);
                    }}
                  />
                </div>
                <div>
                  <label className="text-sm text-slate-700">To date</label>
                  <input
                    type="date"
                    className="mt-1 w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                    value={rentalTo}
                    onChange={(e) => {
                      setError(null);
                      setRentalTo(e.target.value);
                    }}
                  />
                </div>
              </div>
              <div>
                <label className="text-sm text-slate-700">Pickup point (optional)</label>
                <input
                  className="mt-1 w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                  placeholder="Where you will collect the car"
                  value={pickupPoint}
                  onChange={(e) => {
                    setError(null);
                    setPickupPoint(e.target.value);
                  }}
                />
              </div>
              <div>
                <label className="text-sm text-slate-700">Drop-off point (optional)</label>
                <input
                  className="mt-1 w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                  placeholder="Where you will return the car to the owner"
                  value={dropPoint}
                  onChange={(e) => {
                    setError(null);
                    setDropPoint(e.target.value);
                  }}
                />
              </div>
              <div>
                <label className="text-sm text-slate-700">Note (optional)</label>
                <textarea
                  className="mt-1 w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                  rows={2}
                  placeholder="Anything the owner should know"
                  value={customerNote}
                  onChange={(e) => {
                    setError(null);
                    setCustomerNote(e.target.value);
                  }}
                />
              </div>
            </div>
            <div className="mt-6 flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
              <button
                type="button"
                onClick={closeBookingModal}
                className="rounded-md border border-slate-300 px-4 py-2 text-sm hover:bg-slate-50"
              >
                Cancel
              </button>
              <button
                type="button"
                disabled={submitLoading}
                onClick={() => void submitBooking()}
                className="inline-flex items-center justify-center gap-2 rounded-md bg-slate-900 px-4 py-2 text-sm text-white hover:bg-slate-800 disabled:opacity-60"
              >
                {submitLoading ? (
                  <>
                    <ButtonCarSpinner className="text-emerald-200" />
                    Creating…
                  </>
                ) : (
                  "Create booking"
                )}
              </button>
            </div>
          </div>
        </div>
      )}
    </main>
  );
}

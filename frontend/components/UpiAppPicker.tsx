"use client";

/** Popular UPI apps shown for demo checkout and as a preview before Razorpay opens. */
export const UPI_APPS = [
  { id: "gpay", name: "Google Pay", short: "GPay", swatch: "bg-blue-600" },
  { id: "phonepe", name: "PhonePe", short: "Pe", swatch: "bg-violet-600" },
  { id: "paytm", name: "Paytm", short: "Pt", swatch: "bg-sky-500" },
  { id: "bhim", name: "BHIM", short: "BH", swatch: "bg-emerald-600" },
  { id: "amazonpay", name: "Amazon Pay", short: "Am", swatch: "bg-amber-600" },
  { id: "cred", name: "CRED", short: "Cr", swatch: "bg-slate-800" },
] as const;

export type UpiAppId = (typeof UPI_APPS)[number]["id"];

type UpiAppPickerProps = {
  selected: string | null;
  onSelect: (appId: UpiAppId) => void;
  disabled?: boolean;
  subtitle?: string;
};

export function UpiAppPicker({ selected, onSelect, disabled, subtitle }: UpiAppPickerProps) {
  return (
    <div>
      <p className="text-sm text-slate-600">
        {subtitle ?? "Choose the UPI app installed on your phone — you will complete payment in that app."}
      </p>
      <div className="mt-4 grid grid-cols-2 gap-3 sm:grid-cols-3">
        {UPI_APPS.map((app) => {
          const active = selected === app.id;
          return (
            <button
              key={app.id}
              type="button"
              disabled={disabled}
              onClick={() => onSelect(app.id)}
              className={`flex flex-col items-center gap-2 rounded-lg border-2 p-3 text-center transition-colors disabled:opacity-60 ${
                active
                  ? "border-slate-900 bg-slate-50 ring-1 ring-slate-900/10"
                  : "border-slate-200 bg-white hover:border-slate-300 hover:bg-slate-50/80"
              }`}
            >
              <span
                className={`flex h-11 w-11 items-center justify-center rounded-full text-xs font-bold text-white ${app.swatch}`}
                aria-hidden
              >
                {app.short}
              </span>
              <span className="text-xs font-medium leading-tight text-slate-900">{app.name}</span>
            </button>
          );
        })}
      </div>
    </div>
  );
}

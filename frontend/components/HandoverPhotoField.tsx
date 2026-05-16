"use client";

import { useMemo, useRef, useState } from "react";
import { apiForm } from "@/lib/api";
import type { Booking, HandoverPhoto } from "@/lib/apitypes";

export type HandoverPhotoStep =
  | "owner_pickup"
  | "customer_pickup"
  | "customer_return"
  | "owner_return";

type Props = {
  bookingId?: string;
  step: HandoverPhotoStep;
  photos: HandoverPhoto[];
  disabled?: boolean;
  pendingRef: React.MutableRefObject<File[]>;
};

export function HandoverPhotoField({ step, photos, disabled, pendingRef }: Props) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [previews, setPreviews] = useState<string[]>([]);

  const stepPhotos = useMemo(() => photos.filter((p) => p.step === step), [photos, step]);

  function onFilesSelected(files: FileList | null) {
    if (!files?.length || disabled) return;
    const next: File[] = [...pendingRef.current];
    const newPreviews: string[] = [...previews];
    for (let i = 0; i < files.length; i++) {
      const f = files[i];
      if (!f.type.startsWith("image/")) continue;
      if (next.length + stepPhotos.length >= 10) break;
      next.push(f);
      newPreviews.push(URL.createObjectURL(f));
    }
    pendingRef.current = next;
    setPreviews(newPreviews);
    if (inputRef.current) inputRef.current.value = "";
  }

  return (
    <div className="sm:col-span-2">
      <label className="block text-sm text-slate-700">
        Photos <span className="font-normal text-slate-500">(optional, up to 10)</span>
        <input
          ref={inputRef}
          type="file"
          accept="image/jpeg,image/png,image/webp"
          multiple
          disabled={disabled}
          className="mt-1 block w-full text-sm text-slate-700 file:mr-3 file:rounded-md file:border-0 file:bg-slate-100 file:px-3 file:py-2 file:text-sm file:font-medium file:text-slate-800 hover:file:bg-slate-200 disabled:opacity-50"
          onChange={(e) => onFilesSelected(e.target.files)}
        />
      </label>
      {(stepPhotos.length > 0 || previews.length > 0) && (
        <ul className="mt-2 flex flex-wrap gap-2">
          {stepPhotos.map((p) => (
            <li key={p.id}>
              <a href={p.blob_url} target="_blank" rel="noopener noreferrer" className="block">
                <img
                  src={p.blob_url}
                  alt=""
                  className="h-16 w-16 rounded-md border border-slate-200 object-cover"
                />
              </a>
            </li>
          ))}
          {previews.map((url, i) => (
            <li key={`pending-${i}`}>
              <img src={url} alt="" className="h-16 w-16 rounded-md border border-dashed border-slate-300 object-cover" />
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}

export async function uploadPendingHandoverPhotos(
  bookingId: string,
  step: HandoverPhotoStep,
  files: File[]
): Promise<Booking | null> {
  let last: Booking | null = null;
  for (const file of files) {
    const fd = new FormData();
    fd.append("step", step);
    fd.append("file", file);
    const res = await apiForm<{ booking: Booking }>(`/bookings/${bookingId}/handover/photos`, fd);
    last = res.booking;
  }
  return last;
}

/** Razorpay Checkout (https://checkout.razorpay.com/v1/checkout.js) */

export type RazorpaySuccessResponse = {
  razorpay_payment_id: string;
  razorpay_order_id: string;
  razorpay_signature: string;
};

type RazorpayCheckoutInstance = {
  open: () => void;
  on: (event: string, handler: (response: { error: { description?: string } }) => void) => void;
};

type RazorpayConstructor = new (options: Record<string, unknown>) => RazorpayCheckoutInstance;

declare global {
  interface Window {
    Razorpay?: RazorpayConstructor;
  }
}

const SCRIPT_SRC = "https://checkout.razorpay.com/v1/checkout.js";

let scriptPromise: Promise<void> | null = null;

export function loadRazorpayScript(): Promise<void> {
  if (typeof window === "undefined") {
    return Promise.reject(new Error("Razorpay is only available in the browser."));
  }
  if (window.Razorpay) return Promise.resolve();
  if (scriptPromise) return scriptPromise;

  scriptPromise = new Promise((resolve, reject) => {
    const existing = document.querySelector(`script[src="${SCRIPT_SRC}"]`);
    if (existing) {
      existing.addEventListener("load", () => resolve());
      existing.addEventListener("error", () => reject(new Error("Failed to load Razorpay.")));
      return;
    }
    const s = document.createElement("script");
    s.src = SCRIPT_SRC;
    s.async = true;
    s.onload = () => resolve();
    s.onerror = () => reject(new Error("Failed to load Razorpay."));
    document.body.appendChild(s);
  });
  return scriptPromise;
}

/** Map app payment method to Razorpay Checkout method flags. */
export function razorpayMethodFlags(method: string): Record<string, boolean> {
  const allFalse = { upi: false, card: false, netbanking: false, wallet: false };
  switch (method) {
    case "UPI":
      return { ...allFalse, upi: true };
    case "CARD":
      return { ...allFalse, card: true };
    case "NET_BANKING":
      return { ...allFalse, netbanking: true };
    case "QR_CODE":
      return { ...allFalse, upi: true };
    default:
      return { upi: true, card: true, netbanking: true, wallet: false };
  }
}

/** True on phones/tablets where UPI Intent can list installed payment apps. */
export function isMobilePaymentDevice(): boolean {
  if (typeof window === "undefined") return false;
  const ua = navigator.userAgent || "";
  return /Android|iPhone|iPad|iPod|Mobile/i.test(ua);
}

/**
 * Razorpay UPI flow:
 * - UPI → intent (opens installed UPI apps on mobile; QR on desktop)
 * - QR_CODE → qr scan
 */
export function razorpayUpiOptions(method: string): Record<string, unknown> | undefined {
  switch (method) {
    case "QR_CODE":
      return { flow: "qr" };
    case "UPI":
      return { flow: "intent" };
    default:
      return undefined;
  }
}

/** Top-level Checkout flags so mobile browsers can deep-link into UPI apps. */
export function razorpayCheckoutExtras(method: string): Record<string, unknown> {
  if (method === "UPI" && isMobilePaymentDevice()) {
    return { webview_intent: true };
  }
  return {};
}

export type OpenRazorpayCheckoutInput = {
  keyId: string;
  orderId: string;
  amountPaise: number;
  currency: string;
  name?: string;
  description?: string;
  prefillEmail?: string;
  prefillContact?: string;
  paymentMethod: string;
};

export function openRazorpayCheckout(
  input: OpenRazorpayCheckoutInput,
): Promise<RazorpaySuccessResponse> {
  return loadRazorpayScript().then(
    () =>
      new Promise((resolve, reject) => {
        const Razorpay = window.Razorpay;
        if (!Razorpay) {
          reject(new Error("Razorpay is not available."));
          return;
        }

        const options: Record<string, unknown> = {
          key: input.keyId,
          amount: input.amountPaise,
          currency: input.currency || "INR",
          name: input.name ?? "CarManage",
          description: input.description ?? "Trip payment",
          order_id: input.orderId,
          method: razorpayMethodFlags(input.paymentMethod),
          handler: (response: RazorpaySuccessResponse) => resolve(response),
          modal: {
            ondismiss: () => reject(new Error("Payment cancelled.")),
          },
        };

        const upiOpts = razorpayUpiOptions(input.paymentMethod);
        if (upiOpts) {
          options.upi = upiOpts;
        }

        Object.assign(options, razorpayCheckoutExtras(input.paymentMethod));

        if (input.prefillEmail || input.prefillContact) {
          options.prefill = {
            email: input.prefillEmail,
            contact: input.prefillContact,
          };
        }

        const rzp = new Razorpay(options);
        rzp.on("payment.failed", (resp) => {
          reject(new Error(resp?.error?.description ?? "Payment failed."));
        });
        rzp.open();
      }),
  );
}

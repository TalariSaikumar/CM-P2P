/** Round to 2 decimal places (INR). */
function r2(n: number): number {
  return Math.round(n * 100) / 100;
}

function fmt(n: number): string {
  return r2(n).toFixed(2);
}

export type PricingRates = {
  customerCommissionPercent: number;
  ownerCommissionPercent: number;
  gstPercent: number;
  depositPercent: number;
};

export const DEFAULT_PRICING_RATES: PricingRates = {
  customerCommissionPercent: 2,
  ownerCommissionPercent: 1.5,
  gstPercent: 18,
  depositPercent: 75,
};

const EXAMPLE_PER_DAY = 200;
const EXAMPLE_DAYS = 2;

export type PricingExampleBreakdown = {
  perDay: number;
  days: number;
  tripRental: number;
  customerPlatformFee: number;
  customerGst: number;
  customerTotal: number;
  depositDue: number;
  balanceAfterDeposit: number;
  ownerPlatformFee: number;
  ownerGst: number;
  ownerEarnings: number;
};

export function buildPricingExample(rates: PricingRates): PricingExampleBreakdown {
  const tripRental = EXAMPLE_PER_DAY * EXAMPLE_DAYS;
  const customerPlatformFee = r2((tripRental * rates.customerCommissionPercent) / 100);
  const customerGst = r2(((tripRental + customerPlatformFee) * rates.gstPercent) / 100);
  const customerTotal = r2(tripRental + customerPlatformFee + customerGst);
  const depositDue = r2((customerTotal * rates.depositPercent) / 100);
  const balanceAfterDeposit = r2(customerTotal - depositDue);
  const ownerPlatformFee = r2((tripRental * rates.ownerCommissionPercent) / 100);
  const ownerGst = r2((tripRental * rates.gstPercent) / 100);
  const ownerEarnings = r2(tripRental - ownerPlatformFee - ownerGst);

  return {
    perDay: EXAMPLE_PER_DAY,
    days: EXAMPLE_DAYS,
    tripRental,
    customerPlatformFee,
    customerGst,
    customerTotal,
    depositDue,
    balanceAfterDeposit,
    ownerPlatformFee,
    ownerGst,
    ownerEarnings,
  };
}

export { fmt };

export function ratesFromPayment(payment?: {
  customer_commission_percent?: number;
  owner_commission_percent?: number;
  gst_percent_on_commission?: number;
  deposit_percent?: number;
}): Partial<PricingRates> | undefined {
  if (!payment) return undefined;
  return {
    customerCommissionPercent: payment.customer_commission_percent,
    ownerCommissionPercent: payment.owner_commission_percent,
    gstPercent: payment.gst_percent_on_commission,
    depositPercent: payment.deposit_percent,
  };
}

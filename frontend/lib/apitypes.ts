export type AirbagDetail = {
  type: string;
  count: number;
};

/** Owner fleet / GET /cars/:id/edit — includes whether the car is booked for the current UTC calendar day. */
export type CarMineRow = Car & { booked_for_current_date: boolean };

export type Car = {
  id: string;
  owner_id: string;
  car_name: string;
  car_model: string;
  car_number: string;
  registration_number: string;
  engine_number: string;
  price_per_hour: string;
  price_per_day: string;
  price_per_km: string;
  location: string;
  is_active: boolean;
  images?: { id: string; url: string; sort_order: number }[];
  created_at: string;
  model_year: number;
  color: string;
  fuel_type: string;
  transmission: string;
  mileage_km: number;
  num_seats: number;
  airbags: boolean;
  airbag_count: number;
  airbag_details?: AirbagDetail[];
  camera_type: string;
  air_conditioning: boolean;
  cruise_control: boolean;
  open_roof: boolean;
  navigation: boolean;
  speakers: boolean;
};

export type UserSummary = {
  id: string;
  email: string;
  role: string;
  full_name: string;
  phone_number: string;
};

export type BookingPostTripItem = {
  label: string;
  amount_inr: string;
};

export type BookingPaymentBreakdown = {
  payment_status: string;
  payment_method?: string;
  paid_at?: string;
  agreed_base_inr: string;
  trip_days?: number;
  customer_commission_percent: number;
  owner_commission_percent: number;
  customer_commission_inr: string;
  owner_commission_inr: string;
  gst_percent_on_commission: number;
  customer_gst_inr: string;
  owner_gst_inr: string;
  customer_total_inr: string;
  owner_net_inr: string;
  platform_commission_total_inr: string;
  /** unpaid_deposit | awaiting_settlement | final_due | paid */
  payment_phase?: string;
  deposit_percent?: number;
  deposit_due_inr?: string;
  deposit_paid_inr?: string;
  deposit_paid_at?: string;
  trip_balance_inr?: string;
  post_trip_charges_inr?: string;
  final_due_inr?: string;
  owner_projected_payout_inr?: string;
  post_trip_items?: BookingPostTripItem[];
  /** razorpay | simulated — when razorpay, use Razorpay Checkout instead of demo card fields */
  checkout_provider?: string;
  razorpay_key_id?: string;
};

export type BookingCancellation = {
  reason: string;
  cancelled_at: string;
  cancelled_by_role: string;
};

export type HandoverPhoto = {
  id: string;
  step: string;
  blob_url: string;
  created_at: string;
};

export type BookingHandover = {
  owner_pickup_odometer_km?: number | null;
  owner_pickup_fuel_percent?: number | null;
  owner_pickup_notes?: string;
  owner_pickup_recorded_at?: string | null;
  pickup_odometer_km?: number | null;
  pickup_fuel_percent?: number | null;
  pickup_notes?: string;
  pickup_recorded_at?: string | null;
  customer_pickup_accepted_at?: string | null;
  return_odometer_km?: number | null;
  return_fuel_percent?: number | null;
  return_notes?: string;
  return_recorded_at?: string | null;
  owner_return_accepted_at?: string | null;
  photos?: HandoverPhoto[];
};

export type BookingReviewRow = {
  party: string;
  rating: number;
  comment: string;
  reviewer: UserSummary;
  created_at: string;
};

export type Booking = {
  id: string;
  car_id: string;
  customer_id: string;
  owner_id: string;
  status: string;
  final_booking_price?: string | null;
  /** True when the customer accepted the owner's current quoted price. */
  customer_price_accepted?: boolean;
  customer_accepted_price_at?: string | null;
  customer_note?: string;
  rental_from: string;
  rental_to: string;
  pickup_point: string;
  drop_point: string;
  created_at: string;
  car: Car;
  customer: UserSummary;
  owner: UserSummary;
  /** Present when the booking is CONFIRMED and an agreed price exists (commission math for customer pay / owner net). */
  payment?: BookingPaymentBreakdown;
  cancellation?: BookingCancellation | null;
  handover?: BookingHandover | null;
  /** Rental lifecycle stage after deposit (e.g. awaiting_owner_handover, on_trip). */
  trip_stage?: string;
  trip_stage_label?: string;
  reviews?: BookingReviewRow[];
};

export type Message = {
  id: string;
  booking_id: string;
  sender_id: string;
  body: string;
  created_at: string;
  sender: UserSummary;
};

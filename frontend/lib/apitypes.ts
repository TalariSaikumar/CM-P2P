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

export type BookingPaymentBreakdown = {
  payment_status: string;
  payment_method?: string;
  paid_at?: string;
  agreed_base_inr: string;
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
};

export type Booking = {
  id: string;
  car_id: string;
  customer_id: string;
  owner_id: string;
  status: string;
  final_booking_price?: string | null;
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
};

export type Message = {
  id: string;
  booking_id: string;
  sender_id: string;
  body: string;
  created_at: string;
  sender: UserSummary;
};

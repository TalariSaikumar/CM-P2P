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
};

export type UserSummary = {
  id: string;
  email: string;
  role: string;
  full_name: string;
  phone_number: string;
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
};

export type Message = {
  id: string;
  booking_id: string;
  sender_id: string;
  body: string;
  created_at: string;
  sender: UserSummary;
};

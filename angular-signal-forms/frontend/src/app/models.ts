export type Canton = "ZH" | "BE" | "LU" | "UR" | "VS" | "TI" | "GE";
export type Delivery = "pickup" | "sled" | "cable-car";

export const CANTONS: Canton[] = ["ZH", "BE", "LU", "UR", "VS", "TI", "GE"];
export const MOUNTAINS = [
  "Rigi",
  "Pilatus",
  "Matterhorn",
  "Jungfrau",
  "Santis",
];

export interface GuestbookModel {
  name: string;
  email: string;
  canton: Canton;
  yodelVolume: number | null;
  fondueReady: boolean;
  visitDate: Date | null;
  alarmTime: string;
  postcard: string;
}

export interface PicnicItem {
  name: string;
  quantity: number | null;
}

export interface PicnicModel {
  title: string;
  items: PicnicItem[];
}

export interface PermitModel {
  profile: {
    firstName: string;
    lastName: string;
    username: string;
    email: string;
    password: string;
    confirmPassword: string;
  };
  trip: {
    canton: Canton;
    mountain: string;
    altitude: number | null;
    startDate: string;
    delivery: Delivery;
    needsGuide: boolean;
    guideNote: string;
    couponCode: string;
    permitId: string;
  };
  safety: {
    acceptedRules: boolean;
    emergencyPhone: string;
  };
}

export interface AdvancedDeskModel {
  lead: {
    firstName: string;
    lastName: string;
  };
  swissPass: string;
  souvenirBudget: number;
  satisfaction: number;
  cowbellInsurance: boolean;
}

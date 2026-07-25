import { JsonPipe } from "@angular/common";
import { Component, computed, effect, resource, signal } from "@angular/core";
import {
  applyWhen,
  debounce,
  disabled,
  email,
  form,
  FormField,
  hidden,
  max,
  min,
  minLength,
  pattern,
  readonly,
  required,
  submit,
  validate,
  validateAsync,
  validateHttp,
  validateTree,
} from "@angular/forms/signals";
import { Errors } from "./errors";
import { CANTONS, MOUNTAINS, type Canton, type PermitModel } from "./models";

const VALID_COUPON_CODES = new Set(["ALPINE10", "SUMMIT20"]);

function abortableDelay(milliseconds: number, abortSignal: AbortSignal) {
  return new Promise<boolean>((resolve) => {
    if (abortSignal.aborted) {
      resolve(false);
      return;
    }
    const onAbort = () => {
      clearTimeout(timer);
      resolve(false);
    };
    const timer = setTimeout(() => {
      abortSignal.removeEventListener("abort", onAbort);
      resolve(true);
    }, milliseconds);
    abortSignal.addEventListener("abort", onAbort, { once: true });
  });
}

@Component({
  selector: "app-permit-demo",
  imports: [Errors, FormField, JsonPipe],
  templateUrl: "./permit-demo.html",
})
export class PermitDemo {
  readonly cantons: Canton[] = CANTONS;
  readonly mountains = MOUNTAINS;

  readonly permitModel = signal<PermitModel>({
    profile: {
      firstName: "",
      lastName: "",
      username: "fonduepilot",
      email: "",
      password: "",
      confirmPassword: "",
    },
    trip: {
      canton: "VS",
      mountain: "Matterhorn",
      altitude: 3200,
      startDate: "2026-09-12",
      delivery: "cable-car",
      needsGuide: true,
      guideNote: "",
      couponCode: "",
      permitId: "CH-ALP-2026-0001",
    },
    safety: {
      acceptedRules: false,
      emergencyPhone: "",
    },
  });

  readonly permitForm = form(this.permitModel, (path) => {
    required(path.profile.firstName, {
      message: "Enter your first name.",
    });
    required(path.profile.lastName, {
      message: "Enter your last name.",
    });
    required(path.profile.username, {
      message: "Enter a username.",
    });
    debounce(path.profile.username, 350);
    validateHttp(path.profile.username, {
      request: ({ value }) => `/api/usernames/${encodeURIComponent(value())}`,
      onSuccess: (available: { available: boolean }) =>
        available.available
          ? null
          : {
              kind: "usernameTaken",
              message: "This username is already taken.",
            },
      onError: () => ({
        kind: "usernameCheckFailed",
        message: "The username could not be checked. Try again.",
      }),
    });
    required(path.profile.email, {
      message: "Enter your email address.",
    });
    email(path.profile.email, {
      message: "Enter a valid email address.",
    });
    minLength(path.profile.password, 8, {
      message: "Use at least 8 characters.",
    });
    validate(path.profile.confirmPassword, ({ value, valueOf }) =>
      value() === valueOf(path.profile.password)
        ? null
        : {
            kind: "passwordMismatch",
            message: "The passwords do not match.",
          },
    );

    required(path.trip.mountain, {
      message: "Select a mountain.",
    });
    required(path.trip.altitude, {
      message: "Enter an altitude.",
    });
    min(path.trip.altitude, 400, {
      message: "Altitude must be at least 400 meters.",
    });
    max(path.trip.altitude, 4808, {
      message: "Altitude must not exceed 4808 meters.",
    });
    pattern(path.safety.emergencyPhone, /^(\+41|0)[0-9 ]{9,14}$/, {
      message: "Use a Swiss-looking phone number, e.g., +41 44 123 45 67.",
    });
    pattern(path.trip.couponCode, /^[A-Z]+[0-9]+$/, {
      message: "Use uppercase letters followed by digits.",
    });
    validateAsync(path.trip.couponCode, {
      when: ({ value }) => value().trim().length > 0,
      params: ({ value }) => value().trim().toUpperCase(),
      debounce: 300,
      factory: (coupon) =>
        resource({
          params: () => coupon(),
          loader: async ({ params, abortSignal }) => {
            const completed = await abortableDelay(400, abortSignal);
            return completed
              ? { accepted: VALID_COUPON_CODES.has(params) }
              : undefined;
          },
        }),
      onSuccess: (result) =>
        result.accepted
          ? null
          : {
              kind: "couponRejected",
              message: "This coupon code is not accepted.",
            },
      onError: () => ({
        kind: "couponCheckFailed",
        message: "The coupon code could not be checked. Try again.",
      }),
    });
    required(path.safety.acceptedRules, {
      message: "Accept the safety rules.",
    });

    hidden(path.trip.guideNote, {
      when: ({ valueOf }) => !valueOf(path.trip.needsGuide),
    });
    applyWhen(
      path.trip,
      ({ valueOf }) => valueOf(path.trip.needsGuide),
      (trip) => {
        required(trip.guideNote, {
          message: "Enter a note for the guide.",
        });
      },
    );
    disabled(path.trip.couponCode, {
      when: ({ valueOf }) =>
        (valueOf(path.trip.altitude) ?? 0) < 2000
          ? "Coupons are available for trips at or above 2,000 meters."
          : false,
    });
    readonly(path.trip.permitId);
    validateTree(path.profile, ({ valueOf, fieldTree }) => {
      const first = valueOf(path.profile.firstName).trim().toLowerCase();
      const last = valueOf(path.profile.lastName).trim().toLowerCase();
      return first && first === last
        ? {
            kind: "sameName",
            message: "First and last name cannot be identical.",
            fieldTree: fieldTree.lastName,
          }
        : null;
    });
  });

  readonly passwordStrength = computed(() => {
    const password = this.permitForm.profile.password().value();
    const score =
      Number(password.length >= 8) +
      Number(/[A-Z]/.test(password)) +
      Number(/\d/.test(password));
    return ["weak", "steady", "summit-ready"][Math.min(score, 2)];
  });

  readonly draftSavedAt = signal("not yet");
  readonly submitted = signal<PermitModel | null>(null);

  constructor() {
    effect(() => {
      this.permitModel();
      this.draftSavedAt.set(new Date().toLocaleTimeString());
    });
  }

  submitPermit(event: Event): void {
    event.preventDefault();
    submit(this.permitForm, {
      action: async () => {
        await new Promise((resolve) => setTimeout(resolve, 350));
        this.submitted.set(this.permitModel());
      },
      ignoreValidators: "none",
    });
  }

  resetPermit(): void {
    this.permitForm().reset();
    this.submitted.set(null);
  }
}

import { JsonPipe } from "@angular/common";
import { Component, signal } from "@angular/core";
import {
  email,
  form,
  FormField,
  max,
  maxDate,
  maxLength,
  min,
  minDate,
  required,
} from "@angular/forms/signals";
import { Errors } from "./errors";
import { CANTONS, type Canton, type GuestbookModel } from "./models";

@Component({
  selector: "app-guestbook-demo",
  imports: [Errors, FormField, JsonPipe],
  templateUrl: "./guestbook-demo.html",
})
export class GuestbookDemo {
  readonly cantons: Canton[] = CANTONS;

  readonly guestbookModel = signal<GuestbookModel>({
    name: "Heidi Debugger",
    email: "heidi@example.ch",
    canton: "LU",
    yodelVolume: 4,
    fondueReady: true,
    visitDate: new Date("2026-08-01T00:00:00.000Z"),
    alarmTime: "07:15",
    postcard: "The marmot reviewed my pull request.",
  });

  readonly guestbookForm = form(this.guestbookModel, (path) => {
    required(path.name, {
      message: "Enter a guest name.",
    });
    email(path.email, {
      message: "Enter a valid email address.",
    });
    min(path.yodelVolume, 1, {
      message: "Yodel volume must be at least 1.",
    });
    max(path.yodelVolume, 10, {
      message: "Yodel volume must not exceed 10.",
    });
    required(path.visitDate, {
      message: "Select a visit date.",
    });
    minDate(path.visitDate, new Date("2026-07-01T00:00:00.000Z"), {
      message: "The guestbook opens on 1 July 2026.",
    });
    maxDate(path.visitDate, new Date("2026-09-30T00:00:00.000Z"), {
      message: "The guestbook closes on 30 September 2026.",
    });
    maxLength(path.postcard, 140, {
      message: "The postcard must not exceed 140 characters.",
    });
  });

  loadCelebrityGuest(): void {
    this.guestbookModel.set({
      name: "Roger Template-Federer",
      email: "roger@example.ch",
      canton: "ZH",
      yodelVolume: 7,
      fondueReady: true,
      visitDate: new Date("2026-08-01T00:00:00.000Z"),
      alarmTime: "06:30",
      postcard: "Came for the signals, stayed because the form stayed valid.",
    });
  }
}

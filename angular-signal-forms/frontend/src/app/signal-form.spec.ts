import { Injector, signal } from "@angular/core";
import { TestBed } from "@angular/core/testing";
import { FormControl } from "@angular/forms";
import {
  form,
  maxDate,
  minDate,
  pattern,
  required,
} from "@angular/forms/signals";
import { compatForm, extractValue } from "@angular/forms/signals/compat";
import { describe, expect, it } from "vitest";

describe("Signal Forms", () => {
  it("validates a required field", () => {
    const model = signal({ name: "" });
    const testForm = form(
      model,
      (path) => {
        required(path.name, { message: "Enter a name." });
      },
      { injector: TestBed.inject(Injector) },
    );

    expect(testForm.name().getError("required")).toBeDefined();

    testForm.name().value.set("Ada Alpine");

    expect(testForm.name().valid()).toBe(true);
  });

  it("validates date and pattern constraints", () => {
    const model = signal({
      visitDate: new Date("2026-06-30T00:00:00.000Z") as Date | null,
      phone: "123",
    });
    const testForm = form(
      model,
      (path) => {
        minDate(path.visitDate, new Date("2026-07-01T00:00:00.000Z"));
        maxDate(path.visitDate, new Date("2026-09-30T00:00:00.000Z"));
        pattern(path.phone, /^(\+41|0)[0-9 ]{9,14}$/);
      },
      { injector: TestBed.inject(Injector) },
    );

    expect(testForm.visitDate().getError("minDate")).toBeDefined();
    expect(testForm.phone().getError("pattern")).toBeDefined();

    testForm.visitDate().value.set(new Date("2026-08-01T00:00:00.000Z"));
    testForm.phone().value.set("+41 44 123 45 67");

    expect(testForm().valid()).toBe(true);

    testForm.visitDate().value.set(new Date("2026-10-01T00:00:00.000Z"));

    expect(testForm.visitDate().getError("maxDate")).toBeDefined();
  });

  it("extracts raw values from a compatibility form", () => {
    const reference = new FormControl("CH-2026-001", { nonNullable: true });
    const model = signal({ applicant: "Ada Alpine", reference });
    const testForm = compatForm(model, {
      injector: TestBed.inject(Injector),
    });

    expect(extractValue(testForm)).toEqual({
      applicant: "Ada Alpine",
      reference: "CH-2026-001",
    });
  });
});

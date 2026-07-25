import { JsonPipe } from "@angular/common";
import { Component, computed, signal } from "@angular/core";
import {
  FormControl,
  FormGroup,
  ReactiveFormsModule,
  Validators,
} from "@angular/forms";
import {
  compatForm,
  extractValue,
  SignalFormControl,
} from "@angular/forms/signals/compat";
import { FormField, minLength, required } from "@angular/forms/signals";
import { Errors } from "./errors";

@Component({
  selector: "app-reactive-compat-demo",
  imports: [Errors, FormField, JsonPipe, ReactiveFormsModule],
  templateUrl: "./reactive-compat-demo.html",
})
export class ReactiveCompatDemo {
  readonly legacyReferenceControl = new FormControl("CH-2026-001", {
    nonNullable: true,
    validators: Validators.required,
  });

  readonly compatModel = signal({
    applicant: "Ada Alpine",
    reference: this.legacyReferenceControl,
  });

  readonly compatTree = compatForm(this.compatModel, (path) => {
    required(path.applicant, { message: "Enter an applicant name." });
  });

  readonly compatValue = computed(() => extractValue(this.compatTree));
  readonly dirtyCompatValue = computed(() =>
    extractValue(this.compatTree, { dirty: true, enabled: true }),
  );

  readonly bridgeNameControl = new SignalFormControl("Ada Alpine", (path) => {
    required(path, { message: "Enter a name." });
    minLength(path, 3, { message: "Use at least 3 characters." });
  });

  readonly bridgeGroup = new FormGroup({
    name: this.bridgeNameControl,
    platform: new FormControl("Reactive Forms component"),
  });
}

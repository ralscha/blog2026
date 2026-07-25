import { JsonPipe } from "@angular/common";
import { Component, signal } from "@angular/core";
import {
  apply,
  form,
  FormField,
  FormRoot,
  max,
  metadata,
  min,
  required,
  validateStandardSchema,
} from "@angular/forms/signals";
import {
  FIELD_HELP,
  swissPassSchema,
  travelerNameSchema,
} from "./advanced-schemas";
import { ChfInput, CowbellToggle, SummitRating } from "./custom-controls";
import { Errors } from "./errors";
import type { AdvancedDeskModel } from "./models";

@Component({
  selector: "app-advanced-desk-demo",
  imports: [
    ChfInput,
    CowbellToggle,
    Errors,
    FormField,
    FormRoot,
    JsonPipe,
    SummitRating,
  ],
  templateUrl: "./advanced-desk-demo.html",
})
export class AdvancedDeskDemo {
  readonly deskSubmitted = signal<AdvancedDeskModel | null>(null);
  readonly showDeskErrors = signal(false);
  readonly deskModel = signal<AdvancedDeskModel>({
    lead: {
      firstName: "",
      lastName: "",
    },
    swissPass: "CH-1234",
    souvenirBudget: 120,
    satisfaction: 3,
    cowbellInsurance: false,
  });

  readonly deskForm = form(
    this.deskModel,
    (path) => {
      apply(path.lead, travelerNameSchema);
      validateStandardSchema(path, swissPassSchema);
      min(path.souvenirBudget, 20, {
        message: "The souvenir budget must be at least CHF 20.",
      });
      max(path.souvenirBudget, 500, {
        message: "The souvenir budget must not exceed CHF 500.",
      });
      min(path.satisfaction, 4, {
        message: "Select a rating of 4 or 5.",
      });
      required(path.cowbellInsurance, {
        message: "Accept the cowbell insurance option.",
      });
      metadata(
        path.souvenirBudget,
        FIELD_HELP,
        () => "Custom metadata: recommended range of CHF 20–500.",
      );
      metadata(
        path.satisfaction,
        FIELD_HELP,
        () => "Custom metadata: this rating uses a FormValueControl.",
      );
    },
    {
      name: "advanced-desk",
      submission: {
        action: async () => {
          await new Promise((resolve) => setTimeout(resolve, 250));
          this.deskSubmitted.set(this.deskModel());
          this.showDeskErrors.set(false);
          return undefined;
        },
        onInvalid: () => {
          this.showDeskErrors.set(true);
          this.focusFirstDeskError();
        },
        ignoreValidators: "none",
      },
    },
  );

  focusFirstDeskError(): void {
    this.deskForm()
      .errorSummary()[0]
      ?.fieldTree()
      .focusBoundControl({ preventScroll: false });
  }

  deskHelp(field: "souvenirBudget" | "satisfaction"): string {
    return (
      this.deskForm[field]().metadata(FIELD_HELP)?.() ?? "No metadata set."
    );
  }

  resetDesk(): void {
    this.deskForm().reset();
    this.deskSubmitted.set(null);
    this.showDeskErrors.set(false);
  }
}

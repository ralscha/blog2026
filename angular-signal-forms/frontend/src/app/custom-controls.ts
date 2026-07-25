import {
  Component,
  ElementRef,
  input,
  linkedSignal,
  model,
  output,
  viewChild,
} from "@angular/core";
import {
  type FormCheckboxControl,
  type FormValueControl,
} from "@angular/forms/signals";

@Component({
  selector: "app-chf-input",
  templateUrl: "./chf-input.html",
})
export class ChfInput implements FormValueControl<number> {
  readonly value = model.required<number>();
  readonly disabled = input(false);
  readonly invalid = input(false);
  readonly touched = input(false);
  readonly touch = output<void>();
  private readonly inputElement =
    viewChild.required<ElementRef<HTMLInputElement>>("inputElement");

  protected readonly displayValue = linkedSignal(() =>
    this.format(this.value()),
  );

  focus(options?: FocusOptions): void {
    this.inputElement().nativeElement.focus(options);
  }

  reset(): void {
    this.displayValue.set(this.format(this.value()));
  }

  updateModel(): void {
    const normalized = this.displayValue()
      .replace(/['’\s]/g, "")
      .trim();
    const parsed = Number(normalized);
    if (Number.isFinite(parsed)) {
      this.value.set(parsed);
      this.displayValue.set(this.format(parsed));
    } else {
      this.displayValue.set(this.format(this.value()));
    }
  }

  private format(value: number): string {
    return value.toLocaleString("de-CH", { maximumFractionDigits: 2 });
  }
}

@Component({
  selector: "app-summit-rating",
  templateUrl: "./summit-rating.html",
})
export class SummitRating implements FormValueControl<number> {
  readonly value = model.required<number>();
  readonly disabled = input(false);
  readonly invalid = input(false);
  readonly touched = input(false);
  readonly touch = output<void>();
  readonly scores = [1, 2, 3, 4, 5];
  private readonly firstButton =
    viewChild.required<ElementRef<HTMLButtonElement>>("ratingButton");

  focus(options?: FocusOptions): void {
    this.firstButton().nativeElement.focus(options);
  }

  onFocusOut(event: FocusEvent): void {
    const container = event.currentTarget as HTMLElement;
    const nextTarget = event.relatedTarget as Node | null;
    if (!nextTarget || !container.contains(nextTarget)) {
      this.touch.emit();
    }
  }
}

@Component({
  selector: "app-cowbell-toggle",
  templateUrl: "./cowbell-toggle.html",
})
export class CowbellToggle implements FormCheckboxControl {
  readonly checked = model.required<boolean>();
  readonly disabled = input(false);
  readonly touch = output<void>();
  private readonly button =
    viewChild.required<ElementRef<HTMLButtonElement>>("toggleButton");

  focus(options?: FocusOptions): void {
    this.button().nativeElement.focus(options);
  }
}

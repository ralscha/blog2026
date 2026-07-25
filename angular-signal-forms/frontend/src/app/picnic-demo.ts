import { DecimalPipe } from "@angular/common";
import { Component, computed, signal } from "@angular/core";
import {
  applyEach,
  form,
  FormField,
  min,
  minLength,
  required,
} from "@angular/forms/signals";
import { Errors } from "./errors";
import type { PicnicModel } from "./models";

@Component({
  selector: "app-picnic-demo",
  imports: [DecimalPipe, Errors, FormField],
  templateUrl: "./picnic-demo.html",
})
export class PicnicDemo {
  readonly picnicModel = signal<PicnicModel>({
    title: "Emergency fondue logistics",
    items: [
      { name: "Gruyere", quantity: 2 },
      { name: "Bread cube", quantity: 24 },
    ],
  });

  readonly picnicForm = form(this.picnicModel, (path) => {
    required(path.title, {
      message: "Enter a picnic title.",
    });
    minLength(path.items, 1, {
      message: "Add at least one picnic item.",
    });
    applyEach(path.items, (item) => {
      required(item.name, {
        message: "Enter an item name.",
      });
      required(item.quantity, {
        message: "Enter a quantity.",
      });
      min(item.quantity, 1, {
        message: "Quantity must be at least 1.",
      });
    });
  });

  readonly picnicTotal = computed(() =>
    this.picnicModel().items.reduce(
      (sum, item) => sum + Number(item.quantity || 0),
      0,
    ),
  );

  addPicnicItem(): void {
    this.picnicModel.update((model) => ({
      ...model,
      items: [
        ...model.items,
        {
          name: "Chocolate square",
          quantity: 4,
        },
      ],
    }));
    this.picnicForm.items().markAsDirty();
  }

  removePicnicItem(index: number): void {
    this.picnicModel.update((model) => ({
      ...model,
      items: model.items.filter((_, itemIndex) => itemIndex !== index),
    }));
    this.picnicForm.items().markAsDirty();
    this.picnicForm.items().markAsTouched({ skipDescendants: true });
  }
}

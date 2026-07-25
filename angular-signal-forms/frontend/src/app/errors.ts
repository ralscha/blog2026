import { Component, Input } from "@angular/core";

type FieldLike = () => {
  touched(): boolean;
  invalid(): boolean;
  errors(): readonly { message?: string; kind: string }[];
};

@Component({
  selector: "app-errors",
  templateUrl: "./errors.html",
})
export class Errors {
  @Input({ required: true }) field!: FieldLike;
}

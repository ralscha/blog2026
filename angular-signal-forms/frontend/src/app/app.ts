import { Component, ViewEncapsulation } from "@angular/core";
import { AdvancedDeskDemo } from "./advanced-desk-demo";
import { GuestbookDemo } from "./guestbook-demo";
import { PermitDemo } from "./permit-demo";
import { PicnicDemo } from "./picnic-demo";
import { ReactiveCompatDemo } from "./reactive-compat-demo";

@Component({
  selector: "app-root",
  imports: [
    AdvancedDeskDemo,
    GuestbookDemo,
    PermitDemo,
    PicnicDemo,
    ReactiveCompatDemo,
  ],
  templateUrl: "./app.html",
  styleUrl: "./app.css",
  encapsulation: ViewEncapsulation.None,
})
export class App {}

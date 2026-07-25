import { provideHttpClient } from "@angular/common/http";
import { provideSignalFormsConfig } from "@angular/forms/signals";
import { NG_STATUS_CLASSES } from "@angular/forms/signals/compat";
import { bootstrapApplication } from "@angular/platform-browser";
import { App } from "./app/app";

bootstrapApplication(App, {
  providers: [
    provideHttpClient(),
    provideSignalFormsConfig({ classes: NG_STATUS_CLASSES }),
  ],
}).catch((error: unknown) => console.error(error));

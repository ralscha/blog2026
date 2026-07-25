import { createMetadataKey, required, schema } from "@angular/forms/signals";
import type { AdvancedDeskModel } from "./models";

export const FIELD_HELP = createMetadataKey<string>();

export const travelerNameSchema = schema<{
  firstName: string;
  lastName: string;
}>((path) => {
  required(path.firstName, {
    message: "Enter the lead traveler's first name.",
  });
  required(path.lastName, {
    message: "Enter the lead traveler's last name.",
  });
});

export const swissPassSchema = {
  "~standard": {
    version: 1,
    vendor: "tiny-demo-schema",
    validate(value: unknown) {
      const model = value as AdvancedDeskModel;
      const issues: { message: string; path: PropertyKey[] }[] = [];
      if (!/^CH-[0-9]{4}$/.test(model.swissPass)) {
        issues.push({
          message: "Swiss pass must use the format CH-1234.",
          path: ["swissPass"],
        });
      }
      return issues.length ? { issues } : { value: model };
    },
  },
} as const;

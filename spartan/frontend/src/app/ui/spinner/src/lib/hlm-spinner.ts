import { ChangeDetectionStrategy, Component, input } from '@angular/core';
import {
  LucideDynamicIcon,
  LucideLoader2,
  provideLucideIcons,
  type LucideIconInput,
} from '@lucide/angular';
import { classes } from '@spartan-ng/helm/utils';

@Component({
  selector: 'hlm-spinner',
  imports: [LucideDynamicIcon],
  providers: [provideLucideIcons(LucideLoader2)],
  changeDetection: ChangeDetectionStrategy.OnPush,
  host: {
    role: 'status',
    '[attr.aria-label]': 'ariaLabel()',
  },
  template: ` <svg [lucideIcon]="icon()"></svg> `,
})
export class HlmSpinner {
  /**
   * The icon to be used as the spinner.
   * Use provideLucideIcons(...) to register icons referenced by name.
   */
  public readonly icon = input<LucideIconInput>(LucideLoader2);

  /** Aria label for the spinner for accessibility. */
  public readonly ariaLabel = input<string>('Loading', { alias: 'aria-label' });

  constructor() {
    classes(() => 'inline-flex text-[length:--spacing(4)] motion-safe:animate-spin');
  }
}

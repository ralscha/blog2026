import type { BooleanInput, NumberInput } from '@angular/cdk/coercion';
import {
  ChangeDetectionStrategy,
  Component,
  booleanAttribute,
  computed,
  input,
  numberAttribute,
} from '@angular/core';
import {
  LucideCircleCheck,
  LucideInfo,
  LucideLoader2,
  LucideOctagonX,
  LucideTriangleAlert,
} from '@lucide/angular';
import { BrnSonnerImports, type ToasterProps } from '@spartan-ng/brain/sonner';
import { hlm } from '@spartan-ng/helm/utils';
import type { ClassValue } from 'clsx';

@Component({
  selector: 'hlm-toaster',
  imports: [
    BrnSonnerImports,
    LucideCircleCheck,
    LucideInfo,
    LucideLoader2,
    LucideOctagonX,
    LucideTriangleAlert,
  ],
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <brn-sonner-toaster
      [class]="_computedClass()"
      [invert]="invert()"
      [theme]="theme()"
      [position]="position()"
      [hotKey]="hotKey()"
      [richColors]="richColors()"
      [expand]="expand()"
      [duration]="duration()"
      [visibleToasts]="visibleToasts()"
      [closeButton]="closeButton()"
      [toastOptions]="_computedToastOptions()"
      [offset]="offset()"
      [style]="userStyle()"
    >
      <ng-template #loadingIcon>
        <svg lucideLoader2 class="overflow-visible! size-4 motion-safe:animate-spin"></svg>
      </ng-template>
      <ng-template #successIcon>
        <svg lucideCircleCheck class="overflow-visible! size-4"></svg>
      </ng-template>
      <ng-template #errorIcon>
        <svg lucideOctagonX class="overflow-visible! size-4"></svg>
      </ng-template>
      <ng-template #infoIcon>
        <svg lucideInfo class="overflow-visible! size-4"></svg>
      </ng-template>
      <ng-template #warningIcon>
        <svg lucideTriangleAlert class="overflow-visible! size-4"></svg>
      </ng-template>
    </brn-sonner-toaster>
  `,
})
export class HlmToaster {
  public readonly invert = input<ToasterProps['invert'], BooleanInput>(false, {
    transform: booleanAttribute,
  });
  public readonly theme = input<ToasterProps['theme']>('light');
  public readonly position = input<ToasterProps['position']>('bottom-right');
  public readonly hotKey = input<ToasterProps['hotkey']>(['altKey', 'KeyT']);
  public readonly richColors = input<ToasterProps['richColors'], BooleanInput>(false, {
    transform: booleanAttribute,
  });
  public readonly expand = input<ToasterProps['expand'], BooleanInput>(false, {
    transform: booleanAttribute,
  });
  public readonly duration = input<ToasterProps['duration'], NumberInput>(4000, {
    transform: numberAttribute,
  });
  public readonly visibleToasts = input<ToasterProps['visibleToasts'], NumberInput>(3, {
    transform: numberAttribute,
  });
  public readonly closeButton = input<ToasterProps['closeButton'], BooleanInput>(false, {
    transform: booleanAttribute,
  });
  public readonly toastOptions = input<ToasterProps['toastOptions']>({});

  protected readonly _computedToastOptions = computed(() => {
    const options = this.toastOptions();
    return {
      ...options,
      classes: {
        ...options?.classes,
        toast: hlm('rounded-2xl!', options?.classes?.toast),
      },
    };
  });
  public readonly offset = input<ToasterProps['offset']>(null);
  public readonly userClass = input<ClassValue>('', { alias: 'class' });
  public readonly userStyle = input<Record<string, string>>(
    {
      '--normal-bg': 'var(--popover)',
      '--normal-text': 'var(--popover-foreground)',
      '--normal-border': 'var(--border)',
      '--border-radius': 'var(--radius)',
    },
    { alias: 'style' },
  );

  protected readonly _computedClass = computed(() => hlm('toaster group', this.userClass()));
}

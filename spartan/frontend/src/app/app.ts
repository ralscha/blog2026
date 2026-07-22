import { Component, computed, inject, signal } from '@angular/core';
import { FormField, FormRoot, form, required } from '@angular/forms/signals';
import {
  LucideActivity,
  LucideArrowDownRight,
  LucideArrowUpRight,
  LucideBell,
  LucideCircleAlert,
  LucideLayoutDashboard,
  LucideMoon,
  LucidePlus,
  LucideRefreshCw,
  LucideSettings2,
  LucideSun,
} from '@lucide/angular';
import { toast } from '@spartan-ng/brain/sonner';
import { HlmAlertImports } from '@spartan-ng/helm/alert';
import { HlmAvatarImports } from '@spartan-ng/helm/avatar';
import { HlmBadgeImports } from '@spartan-ng/helm/badge';
import { HlmButtonImports } from '@spartan-ng/helm/button';
import { HlmCardImports } from '@spartan-ng/helm/card';
import { HlmDialogImports } from '@spartan-ng/helm/dialog';
import { HlmFieldImports } from '@spartan-ng/helm/field';
import { HlmInputImports } from '@spartan-ng/helm/input';
import { HlmProgressImports } from '@spartan-ng/helm/progress';
import { HlmSeparatorImports } from '@spartan-ng/helm/separator';
import { HlmSpinnerImports } from '@spartan-ng/helm/spinner';
import { HlmSwitchImports } from '@spartan-ng/helm/switch';
import { HlmTableImports } from '@spartan-ng/helm/table';
import { HlmTabsImports } from '@spartan-ng/helm/tabs';
import { HlmToasterImports } from '@spartan-ng/helm/sonner';
import { HlmToggleGroupImports } from '@spartan-ng/helm/toggle-group';
import { HlmTooltipImports } from '@spartan-ng/helm/tooltip';
import { DashboardApi } from './dashboard-api.service';

@Component({
  selector: 'app-root',
  imports: [
    FormField,
    FormRoot,
    LucideActivity,
    LucideArrowDownRight,
    LucideArrowUpRight,
    LucideBell,
    LucideCircleAlert,
    LucideLayoutDashboard,
    LucideMoon,
    LucidePlus,
    LucideRefreshCw,
    LucideSettings2,
    LucideSun,
    ...HlmAlertImports,
    ...HlmAvatarImports,
    ...HlmBadgeImports,
    ...HlmButtonImports,
    ...HlmCardImports,
    ...HlmDialogImports,
    ...HlmFieldImports,
    ...HlmInputImports,
    ...HlmProgressImports,
    ...HlmSeparatorImports,
    ...HlmSpinnerImports,
    ...HlmSwitchImports,
    ...HlmTableImports,
    ...HlmTabsImports,
    ...HlmToasterImports,
    ...HlmToggleGroupImports,
    ...HlmTooltipImports,
  ],
  templateUrl: './app.html',
  styleUrl: './app.css',
})
export class App {
  protected readonly api = inject(DashboardApi);
  protected readonly darkMode = signal(false);
  protected readonly dailyDigest = signal(true);
  protected readonly density = signal('comfortable');
  protected readonly projectModel = signal({ name: '' });
  protected readonly projectForm = form(this.projectModel, (path) => {
    required(path.name, { message: 'Give your project a name.' });
  });
  protected readonly onTrackProjects = computed(
    () =>
      this.api.dashboard()?.projects.filter((project) => project.status === 'On track').length ?? 0,
  );

  constructor() {
    this.api.load();
  }

  protected refresh(): void {
    this.api.load();
    toast.info('Refreshing workspace data');
  }

  protected toggleTheme(): void {
    this.darkMode.update((enabled) => !enabled);
    document.documentElement.classList.toggle('dark', this.darkMode());
  }

  protected createProject(event: SubmitEvent, dialog: { close(): void }): void {
    event.preventDefault();
    const name = this.projectModel().name.trim();
    if (!name) return;
    this.api.createProject(name, () => {
      dialog.close();
      this.projectModel.set({ name: '' });
      toast.success('Project created');
    });
  }

  protected updateDigest(dailyDigest: boolean): void {
    this.dailyDigest.set(dailyDigest);
    this.api.saveDigest(dailyDigest, () => toast.success('Notification preference saved'));
  }

  protected setDensity(density: string | string[] | null | undefined): void {
    if (typeof density === 'string') this.density.set(density);
  }
}

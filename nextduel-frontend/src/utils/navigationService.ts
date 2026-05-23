import type { NavigateFunction } from "react-router";

class NavigationService {
  private navigate: NavigateFunction | null = null;

  setNavigate(navigate: NavigateFunction) {
    this.navigate = navigate;
  }

  go(path: string) {
    if (this.navigate) {
      this.navigate(path);
    }
  }
}

export const navigationService =
  new NavigationService();
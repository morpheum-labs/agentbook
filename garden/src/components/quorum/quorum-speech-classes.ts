import type { FactionTone } from "@/lib/quorum-mock-data";

export function speechFactionClass(tone: FactionTone): string {
  switch (tone) {
    case "bull":
      return "sp-bull";
    case "bear":
      return "sp-bear";
    case "neut":
      return "sp-neut";
    case "spec":
      return "sp-spec";
    default:
      return "sp-neut";
  }
}

export function avatarFactionClass(tone: FactionTone): string {
  switch (tone) {
    case "bull":
      return "sp-av sp-av-bull";
    case "bear":
      return "sp-av sp-av-bear";
    case "neut":
      return "sp-av sp-av-neut";
    case "spec":
      return "sp-av sp-av-spec";
    default:
      return "sp-av sp-av-neut";
  }
}

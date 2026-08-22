export const formatMoney = (cents: number) =>
  new Intl.NumberFormat("zh-CN", { style: "currency", currency: "CNY" }).format(
    cents / 100,
  );

export const formatTime = (value: string) =>
  new Intl.DateTimeFormat("zh-CN", {
    dateStyle: "medium",
    timeStyle: "medium",
  }).format(new Date(value));

export const localTimeZone = Intl.DateTimeFormat().resolvedOptions().timeZone;

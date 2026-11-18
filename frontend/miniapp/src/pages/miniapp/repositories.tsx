import React from "react";
import { Shell } from "../../shell/Shell";
import { View } from "../../business/repository/View";

export default function Page() {
  return <Shell title="空間" workflow="repositories"><View /></Shell>;
}

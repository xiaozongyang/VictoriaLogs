import { FC, useMemo } from "preact/compat";

interface Props {
  data: Record<string, string>[]
}

export const JsonView: FC<Props> = ({ data }) => {
  const jsonStr = useMemo(() => {
    return data.map((a) => JSON.stringify(a, null, 4)).join("\n")
  }, [data]);
  return (
    <pre style="line-height: 1.2em">{jsonStr}</pre>
  );
};

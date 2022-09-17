import Long from "long";

export type NumberLong = {
  low: number;
  high: number;
  unsigned: boolean;
};

function longToDate(l: NumberLong): Date {
  const value = new Long(l.low, l.high, l.unsigned);
  return new Date(value.div(1000000).toNumber());
}

export default longToDate;

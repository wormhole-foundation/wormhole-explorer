import { Box, Card, Typography } from "@mui/material";
import {
  createColumnHelper,
  getCoreRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable,
} from "@tanstack/react-table";
import { useState } from "react";
import useHeartbeats, { HeartbeatResponse } from "../hooks/useHeartbeats";
import longToDate from "../utils/longToDate";
import Table from "./Table";

const columnHelper = createColumnHelper<HeartbeatResponse>();

const columns = [
  columnHelper.accessor("nodename", {
    header: () => "Guardian",
    sortingFn: `text`,
  }),
  columnHelper.accessor("version", {
    header: () => "Version",
  }),
  columnHelper.accessor("features", {
    header: () => "Features",
    cell: (info) => {
      const value = info.getValue();
      return value && value.length > 0 ? value.join(", ") : "none";
    },
  }),
  columnHelper.accessor("counter", {
    header: () => "Counter",
  }),
  columnHelper.accessor("boottimestamp", {
    header: () => "Boot",
    cell: (info) =>
      info.getValue() ? longToDate(info.getValue()).toLocaleString() : null,
  }),
  columnHelper.accessor("timestamp", {
    header: () => "Timestamp",
    cell: (info) =>
      info.getValue() ? longToDate(info.getValue()).toLocaleString() : null,
  }),
  columnHelper.accessor("guardianaddr", {
    header: () => "Address",
    cell: (info) => (
      <Box component="pre" m={0}>
        {info.getValue()}
      </Box>
    ),
  }),
];

function Guardians() {
  const heartbeats = useHeartbeats();
  const [sorting, setSorting] = useState<SortingState>([]);
  const table = useReactTable({
    columns,
    data: heartbeats,
    state: {
      sorting,
    },
    getRowId: (heartbeat) => heartbeat.guardianaddr,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    onSortingChange: setSorting,
  });
  return (
    <Box m={2}>
      <Card>
        <Table<HeartbeatResponse> table={table} />
      </Card>
    </Box>
  );
}

export default Guardians;

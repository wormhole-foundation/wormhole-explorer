import { ChevronRight } from "@mui/icons-material";
import { Card, IconButton, Typography } from "@mui/material";
import { Box } from "@mui/system";
import {
  createColumnHelper,
  getCoreRowModel,
  getExpandedRowModel,
  Row,
  useReactTable,
} from "@tanstack/react-table";
import { ReactElement, useMemo } from "react";
import useLatestObservations, {
  ObservationsResponse,
} from "../hooks/useLatestObservations";
import useLatestVAAs, { VAAsResponse } from "../hooks/useLatestVAAs";
import * as vaa from "../utils/vaa";
import Table from "./Table";

const columnHelper = createColumnHelper<VAAsResponse>();

// TODO: parse once

const columns = [
  columnHelper.display({
    id: "_expand",
    cell: ({ row }) =>
      row.getCanExpand() ? (
        <IconButton
          size="small"
          {...{
            onClick: row.getToggleExpandedHandler(),
            style: { cursor: "pointer" },
          }}
        >
          <ChevronRight
            sx={{
              transition: ".2s",
              transform: row.getIsExpanded() ? "rotate(90deg)" : undefined,
            }}
          />
        </IconButton>
      ) : null,
  }),
  columnHelper.accessor("_id", {
    id: "sequence",
    header: () => "Sequence",
    cell: (info) => info.getValue().split("/")[2],
  }),
  columnHelper.accessor("vaa", {
    id: "type",
    header: () => "Type",
    cell: (info) =>
      vaa.parse(Buffer.from(info.getValue(), "base64")).payload.type,
  }),
  columnHelper.accessor("vaa", {
    id: "chain",
    header: () => "Chain",
    cell: (info) =>
      (vaa.parse(Buffer.from(info.getValue(), "base64")).payload as any)
        .chain || "",
  }),
  columnHelper.accessor("vaa", {
    id: "address",
    header: () => "Address",
    cell: (info) =>
      (vaa.parse(Buffer.from(info.getValue(), "base64")).payload as any)
        .address || "",
  }),
  columnHelper.accessor("vaa", {
    id: "module",
    header: () => "Module",
    cell: (info) =>
      (vaa.parse(Buffer.from(info.getValue(), "base64")).payload as any)
        .module || "",
  }),
  columnHelper.accessor("createdAt", {
    header: () => "Observed At",
    cell: (info) => new Date(info.getValue()).toLocaleString(),
  }),
];

const obsColumnHelper = createColumnHelper<CollatedObservation>();

const obsColumns = [
  obsColumnHelper.accessor("_id", {
    header: () => "Sequence",
    cell: (info) => info.getValue().split("/")[2],
  }),
  obsColumnHelper.accessor("count", {
    header: () => "Count",
    cell: (info) => info.getValue(),
  }),
];

function VAADetails({ row }: { row: Row<VAAsResponse> }): ReactElement {
  return (
    <Typography variant="body2" sx={{ wordBreak: "break-all" }}>
      {Buffer.from(row.original.vaa, "base64").toString("hex")}
    </Typography>
  );
}

type CollatedObservation = {
  _id: string;
  count: number;
  observations: ObservationsResponse[];
};

function Governance() {
  const vaas = useLatestVAAs(
    "1/0000000000000000000000000000000000000000000000000000000000000004"
  );
  const obs = useLatestObservations(
    "1/0000000000000000000000000000000000000000000000000000000000000004"
  );
  const collatedObservations: CollatedObservation[] = useMemo(
    () =>
      // NOTE: this ignores differing digests
      Object.entries(
        obs.reduce((obvsById, o) => {
          if (!obvsById[o.messageid]) {
            obvsById[o.messageid] = [];
          }
          obvsById[o.messageid].push(o);
          return obvsById;
        }, {} as any)
      ).map(([key, val]: [string, any]) => ({
        _id: key,
        count: val.length,
        observations: val,
      })),
    [obs]
  );
  const table = useReactTable({
    columns,
    data: vaas,
    getRowId: (vaa) => vaa._id,
    getRowCanExpand: () => true,
    getCoreRowModel: getCoreRowModel(),
    getExpandedRowModel: getExpandedRowModel(),
    enableSorting: false,
  });
  const obsTable = useReactTable({
    columns: obsColumns,
    data: collatedObservations,
    getRowId: (o) => o._id,
    getCoreRowModel: getCoreRowModel(),
    enableSorting: false,
  });
  return (
    <>
      <Box m={2}>
        <Card>
          <Box m={2}>
            <Typography variant="h5">Latest Governance VAAs</Typography>
            <Table<VAAsResponse>
              table={table}
              renderSubComponent={VAADetails}
            />
          </Box>
        </Card>
      </Box>
      <Box m={2}>
        <Card>
          <Box m={2}>
            <Typography variant="h5">Latest Governance Observations</Typography>
            <Typography variant="caption">
              Collated from the latest 100 governance observations
            </Typography>
            <Table<CollatedObservation> table={obsTable} />
          </Box>
        </Card>
      </Box>
    </>
  );
}
export default Governance;

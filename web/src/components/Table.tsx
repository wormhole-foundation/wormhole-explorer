import { ArrowDownward, ArrowUpward } from "@mui/icons-material";
import {
  Box,
  Collapse,
  Divider,
  SxProps,
  Table as MuiTable,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Theme,
  useTheme,
} from "@mui/material";
import { grey } from "@mui/material/colors";
import { flexRender, Row, Table as TanTable } from "@tanstack/react-table";
import { Fragment, ReactElement } from "react";
import { ErrorBoundary } from "react-error-boundary";
import ErrorFallback from "./ErrorFallback";

function Table<T>({
  table,
  conditionalRowStyle,
  renderSubComponent: SubComponent,
}: {
  table: TanTable<T>;
  conditionalRowStyle?: (a: T) => SxProps<Theme> | undefined;
  renderSubComponent?: ({ row }: { row: Row<T> }) => ReactElement;
}) {
  const theme = useTheme();
  return (
    <TableContainer>
      <MuiTable size="small">
        <TableHead>
          {table.getHeaderGroups().map((headerGroup) => (
            <TableRow key={headerGroup.id}>
              {headerGroup.headers.map((header) => (
                <TableCell
                  key={header.id}
                  sx={
                    header.column.getCanSort()
                      ? {
                          cursor: "pointer",
                          userSelect: "select-none",
                          ":hover": {
                            background:
                              theme.palette.mode === "dark"
                                ? grey[800]
                                : grey[100],
                          },
                        }
                      : {}
                  }
                  onClick={header.column.getToggleSortingHandler()}
                >
                  <Box display="flex" alignContent="center">
                    {header.isPlaceholder
                      ? null
                      : flexRender(
                          header.column.columnDef.header,
                          header.getContext()
                        )}
                    <Box flexGrow={1} />
                    <Box display="flex" alignItems="center">
                      {{
                        asc: <ArrowUpward fontSize="small" sx={{ ml: 0.5 }} />,
                        desc: (
                          <ArrowDownward fontSize="small" sx={{ ml: 0.5 }} />
                        ),
                      }[header.column.getIsSorted() as string] ?? null}
                    </Box>
                  </Box>
                </TableCell>
              ))}
            </TableRow>
          ))}
        </TableHead>
        <TableBody>
          {table.getRowModel().rows.map((row) => (
            <Fragment key={row.id}>
              <TableRow
                sx={
                  conditionalRowStyle ? conditionalRowStyle(row.original) : {}
                }
              >
                {row.getVisibleCells().map((cell) => (
                  <TableCell key={cell.id}>
                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                  </TableCell>
                ))}
              </TableRow>
              {SubComponent && (
                <TableRow>
                  {/* 2nd row is a custom 1 cell row */}
                  <TableCell
                    colSpan={row.getVisibleCells().length}
                    sx={{ p: 0, borderBottom: 0 }}
                  >
                    <Collapse in={row.getIsExpanded()}>
                      <Box p={1.5}>
                        <ErrorBoundary FallbackComponent={ErrorFallback}>
                          <SubComponent row={row} />
                        </ErrorBoundary>
                      </Box>
                      <Divider />
                    </Collapse>
                  </TableCell>
                </TableRow>
              )}
            </Fragment>
          ))}
        </TableBody>
      </MuiTable>
    </TableContainer>
  );
}
export default Table;

import { FC, useMemo, useRef, useState, useEffect, useCallback } from "preact/compat";
import Button from "../../../../../../components/Main/Button/Button";
import { SettingsIcon } from "../../../../../../components/Main/Icons";
import Tooltip from "../../../../../../components/Main/Tooltip/Tooltip";
import Select from "../../../../../../components/Main/Select/Select";
import useBoolean from "../../../../../../hooks/useBoolean";
import Modal from "../../../../../../components/Main/Modal/Modal";
import { useSearchParams } from "react-router-dom";
import "./style.scss";
import { SortDirection } from "../types";

const title = "JSON settings";
const directionList = ["asc", "desc"];

interface JsonSettingsProps {
  fields: string[];
  sortQueryParamName: string;
}

const JsonViewSettings: FC<JsonSettingsProps> = ({
  fields,
  sortQueryParamName
}) => {
  const [searchParams, setSearchParams] = useSearchParams();
  const buttonRef = useRef<HTMLDivElement>(null);

  const {
    value: openSettings,
    toggle: toggleOpenSettings,
    setFalse: handleClose,
  } = useBoolean(false);

  const [sortField, setSortField] = useState<string | null>(null);
  const [sortDirection, setSortDirection] = useState<SortDirection>(null);

  useEffect(() => {
    const sortParam = searchParams.get(sortQueryParamName);
    const isSortDirection = (value: string) : value is Exclude<SortDirection, null> => directionList.includes(value);
    if (sortParam) {
      const [field, direction] = sortParam.split(":").map(decodeURIComponent);
      if (field && (isSortDirection(direction))) {
        setSortField(field);
        setSortDirection(direction);
      }
    }
  }, [searchParams, sortQueryParamName, setSortField, setSortDirection]);

  const updateSortParams = useCallback((field: string | null, direction: SortDirection) => {
    const updatedParams = new URLSearchParams(searchParams.toString());

    if (!field || !direction) {
      updatedParams.delete(sortQueryParamName);
    } else {
      updatedParams.set(sortQueryParamName, `${field}:${direction || ""}`);
    }

    setSearchParams(updatedParams);
  }, [searchParams, sortQueryParamName]);

  const handleSort = (field: string) => {
    const newDirection: SortDirection = sortDirection || "asc";
    setSortField(field);
    setSortDirection(newDirection);
    updateSortParams(field, newDirection);
  };

  const resetSort = () => {
    setSortField(null);
    setSortDirection(null);
    updateSortParams(null, null);
  };

  const handleChangeSortDirection = (direction: string) => {
    const field = sortField || fields[0];
    setSortField(field);
    setSortDirection(direction as SortDirection);
    updateSortParams(field, direction as SortDirection);
  };

  return (
    <div className="vm-json-settings">
      <Tooltip title={title}>
        <div ref={buttonRef}>
          <Button
            variant="text"
            startIcon={<SettingsIcon/>}
            onClick={toggleOpenSettings}
            ariaLabel={title}
          />
        </div>
      </Tooltip>
      {openSettings && (
        <Modal
          title={title}
          className="vm-json-settings-modal"
          onClose={handleClose}
        >
          <div className="vm-json-settings-modal-section">
            <div className="vm-json-settings-modal-section__sort-settings-container">
              <Select
                value={sortField || ""}
                onChange={handleSort}
                list={fields}
                label="Select field"
              />
              <Select
                value={sortDirection || ""}
                onChange={handleChangeSortDirection}
                list={directionList}
                label="Sort direction"
              />
              {(sortField || sortDirection) && (
                <Button
                  variant="outlined"
                  color="error"
                  onClick={resetSort}
                >
                  Reset sort
                </Button>
              )}
            </div>
          </div>
        </Modal>)}
    </div>
  );
};

export default JsonViewSettings;

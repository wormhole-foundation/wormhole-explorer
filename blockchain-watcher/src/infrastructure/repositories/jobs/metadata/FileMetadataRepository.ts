import fs from "fs";
import { MetadataRepository } from "../../../../domain/repositories";

export class FileMetadataRepository implements MetadataRepository<any> {
  private readonly dirPath: string;
  private clone: MetadataRepository<any> | undefined;

  constructor(dirPath: string, clone?: MetadataRepository<any>) {
    this.dirPath = dirPath;
    this.clone = clone;
    if (!fs.existsSync(this.dirPath)) {
      fs.mkdirSync(this.dirPath, { recursive: true });
    }
  }

  async get(id: string): Promise<any> {
    const filePath = `${this.dirPath}/${id}.json`;

    return fs.promises
      .readFile(filePath, "utf8")
      .then(JSON.parse)
      .catch((err) => null);
  }

  async save(id: string, metadata: any): Promise<void> {
    const filePath = `${this.dirPath}/${id}.json`;
    await fs.promises.writeFile(filePath, JSON.stringify(metadata), "utf8");
    if (this.clone) {
      await this.clone.save(id, metadata);
    }
  }
}

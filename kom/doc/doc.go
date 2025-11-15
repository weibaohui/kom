package doc

import (
    "encoding/json"
    "strings"

    "github.com/duke-git/lancet/v2/slice"
    openapi_v2 "github.com/google/gnostic-models/openapiv2"
    "github.com/weibaohui/kom/utils"
    "k8s.io/klog/v2"
)

// 移除全局变量 trees

type Docs struct {
	Trees []TreeNode
}

// TreeNode 表示树形结构的节点
type TreeNode struct {
	ID              string      `json:"id"` // GVK形式io.k8s.apimachinery.pkg.apis.meta.v1.MicroTime
	Label           string      `json:"label"`
	Value           string      `json:"value"` // amis tree 需要
	Description     string      `json:"description,omitempty"`
	Type            string      `json:"type,omitempty"`
	Ref             string      `json:"ref,omitempty"`
	Enum            []Enum      `json:"enum,omitempty"`
	Items           Items       `json:"items,omitempty"`
	VendorExtension interface{} `json:"vendor_extension,omitempty"`
	Children        []*TreeNode `json:"children,omitempty"`
	group           string      // 从ID中尝试解析GVK，方便查询，不一定准确
	version         string      // 从ID中尝试解析GVK，方便查询
	kind            string      // 从ID中尝试解析GVK，方便查询
	FullId          string      `json:"full_id,omitempty"` // 完全ID
}

// SchemaDefinition 表示根定义
type SchemaDefinition struct {
	Name  string      `json:"name"`
	Value SchemaValue `json:"value"`
}

// SchemaValue 表示定义的值
type SchemaValue struct {
	Description     string           `json:"description"`
	Properties      SchemaProperties `json:"properties"`
	Type            SchemaType       `json:"type"`
	VendorExtension []interface{}    `json:"vendor_extension,omitempty"`
}

// SchemaProperties 表示属性
type SchemaProperties struct {
	AdditionalProperties []Property `json:"additional_properties"`
}

// Property 表示单个属性
type Property struct {
	Name  string        `json:"name"`
	Value PropertyValue `json:"value"`
}

// PropertyValue 表示属性的值
type PropertyValue struct {
	Description     string           `json:"description,omitempty"`
	Type            *SchemaType      `json:"type,omitempty"`
	Ref             string           `json:"_ref,omitempty"`
	Enum            []Enum           `json:"enum,omitempty"`
	Items           Items            `json:"items,omitempty"`
	VendorExtension interface{}      `json:"vendor_extension,omitempty"`
	Properties      SchemaProperties `json:"properties"`
}
type Enum struct {
	Yaml string `json:"yaml,omitempty"`
}
type Schema struct {
	Ref string `json:"_ref,omitempty"`
}
type Items struct {
	Schema []Schema `json:"schema,omitempty"`
}

// SchemaType 表示类型
type SchemaType struct {
	Value []string `json:"value,omitempty"`
}

// RootDefinitions 最外层定义
type RootDefinitions struct {
	Swagger     string      `json:"swagger"`
	Definitions Definitions `json:"definitions,omitempty"`
}

// Definitions 表示所有定义
// 使用interface{}
type Definitions struct {
	AdditionalProperties []map[string]interface{} `json:"additional_properties"`
}

// 注意：移除全局 definitionsMap，避免并发写入导致崩溃

var blackList = []string{
    "#/definitions/io.k8s.apiextensions-apiserver.pkg.apis.apiextensions.v1.JSONSchemaProps",
    "#/definitions/io.k8s.apiextensions-apiserver.pkg.apis.apiextensions.v1beta1.JSON",
    "#/definitions/io.k8s.apiextensions-apiserver.pkg.apis.apiextensions.v1beta1.JSONSchemaProps",
}

// parseOpenAPISchema 解析 OpenAPI Schema JSON 字符串并返回根 TreeNode
// 参数:
//   - schemaJSON: OpenAPI 单个 definition 的 JSON 字符串
//   - defs: definitions 映射，用于处理 $ref 引用，采用局部传参避免并发问题
// 返回:
//   - TreeNode: 构建的树根节点
//   - error: 解析错误
// Example:
//
//	  JSON样例
//		 "name": "com.example.stable.v1.CronTab",
//			"value": { },
//			"properties": {
//			    "additional_properties": [ {},{}]
//			  },
//			  "vendor_extension": [ {},{}]
//			}
func parseOpenAPISchema(schemaJSON string, defs map[string]SchemaDefinition) (TreeNode, error) {
    var def SchemaDefinition
    err := json.Unmarshal([]byte(schemaJSON), &def)
    if err != nil {
        return TreeNode{}, err
    }
    // 将解析后的定义写入局部映射，避免全局并发写入
    defs[def.Name] = def
    // 构建树
    return buildTree(def, "", defs), nil
}
func parseID(id string) (group, version, kind string) {
	parts := strings.Split(id, ".")
	if len(parts) < 3 {
		return "", "", "" // 不足三个部分，无法解析
	}

	kind = parts[len(parts)-1]    // 最后一段是 Kind
	version = parts[len(parts)-2] // 倒数第二段是 Version

	if len(parts) > 3 { // 判断是否有 Group 部分
		groupParts := parts[:len(parts)-2] // Group 是前面的部分
		group = groupParts[len(groupParts)-1]
	}
	if group == "core" {
		// core 在书写yaml时不需要写，但是解析出来还是有core，这里做一下处理
		group = ""
	}

	return group, version, kind
}

// buildTree 根据 SchemaDefinition 构建 TreeNode
// 参数:
//   - def: 单个 SchemaDefinition
//   - parentId: 父节点ID，用于构建 FullId
//   - defs: definitions 映射，用于处理 $ref 引用
// 返回:
//   - TreeNode: 构建的树节点
func buildTree(def SchemaDefinition, parentId string, defs map[string]SchemaDefinition) TreeNode {
    // 打印中文日志，便于定位构建过程
    klog.V(8).Infof("构建树节点: %s", def.Name)

	labelParts := strings.Split(def.Name, ".")
	label := labelParts[len(labelParts)-1]

	nodeType := ""
	if len(def.Value.Type.Value) > 0 {
		nodeType = def.Value.Type.Value[0]
	}
	var children []*TreeNode

    for _, prop := range def.Value.Properties.AdditionalProperties {
        children = append(children, buildPropertyNode(prop, def.Name, defs))
    }

	group, version, kind := parseID(def.Name)
	return TreeNode{
		ID:          def.Name,
		FullId:      parentId + "." + def.Name,
		Label:       label,
        Value:       utils.RandNLengthString(20),
		Description: def.Value.Description,
		Type:        nodeType,
		Children:    children,
		group:       group,
		version:     version,
		kind:        kind,
	}

}

// buildPropertyNode 根据 Property 构建 TreeNode
// 参数:
//   - prop: 属性定义
//   - parentId: 父节点ID
//   - defs: definitions 映射，用于解析 $ref 引用
// 返回:
//   - *TreeNode: 构建的属性节点
func buildPropertyNode(prop Property, parentId string, defs map[string]SchemaDefinition) *TreeNode {
    label := prop.Name
    nodeID := prop.Name
    fullID := parentId + "." + prop.Name
    description := prop.Value.Description
    nodeType := ""
    ref := ""

	if prop.Value.Type != nil && len(prop.Value.Type.Value) > 0 {
		nodeType = prop.Value.Type.Value[0]
	}
	if prop.Value.Ref != "" {
		ref = prop.Value.Ref
	}

	var children []*TreeNode

	// 如果有引用，查找定义并递归构建子节点
    if ref != "" && !slice.Contain(blackList, ref) {
        // 假设 ref 的格式为 "#/definitions/io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta"
        refParts := strings.Split(ref, "/")
        refName := refParts[len(refParts)-1]
        // 构建完整的引用路径
        // fullRef := strings.Join(refParts[1:], ".")

        // 这个可能会导致 循环引用溢出
        if def, exists := defs[refName]; exists {
            if !slice.Contain(blackList, refName) {
                childNode := buildTree(def, fullID, defs)
                children = append(children, &childNode)
            }
        } else {
            // 如果引用的定义不存在，可以记录为一个叶子节点或处理为需要进一步扩展
            children = append(children, &TreeNode{
                ID:          refName,
                FullId:      fullID + "." + refName,
                Label:       refName,
                Value:       refName,
                Description: "未找到引用的 definition",
            })
        }
    }

    for _, pp := range prop.Value.Properties.AdditionalProperties {
        children = append(children, buildPropertyNode(pp, fullID, defs))
    }

	return &TreeNode{
		ID:          nodeID,
		FullId:      fullID,
		Label:       label,
		Value:       nodeID,
		Description: description,
		Type:        nodeType,
		Ref:         ref,
		Items:       prop.Value.Items,
		Enum:        prop.Value.Enum,
		Children:    children,
	}
}

// printTree 递归打印 TreeNode
func printTree(node *TreeNode, level int) {
	indent := strings.Repeat("  ", level)
	klog.V(2).Infof("%s%s (ID: %s)\n", indent, node.Label, node.ID)
	if node.Description != "" {
		klog.V(2).Infof("%s  Description: %s\n", indent, node.Description)
	}
	if node.Type != "" {
		klog.V(2).Infof("%s  Type: %s\n", indent, node.Type)
	}
	if node.Ref != "" {
		klog.V(2).Infof("%s  Ref: %s\n", indent, node.Ref)
	}

	for _, child := range node.Children {
		printTree(child, level+1)
	}
}

// InitTrees 解析 OpenAPI Schema 并构建 Docs 结构体，避免全局变量，减少并发风险
// 参数:
//   - schema: OpenAPI v2 文档对象
// 返回:
//   - *Docs: 构建好的文档树
func InitTrees(schema *openapi_v2.Document) *Docs {
    // 使用局部 definitionsMap，避免并发写入导致崩溃
    definitionsMap := make(map[string]SchemaDefinition)

	// 将 OpenAPI Schema 转换为 JSON 字符串
    schemaBytes, err := json.Marshal(schema)
    if err != nil {
        klog.V(2).Infof("序列化 OpenAPI Schema 为 JSON 失败: %v", err)
        return nil
    }

	root := &RootDefinitions{}
    err = json.Unmarshal(schemaBytes, root)
    if err != nil {
        klog.V(2).Infof("反序列化 OpenAPI Schema 失败: %v", err)
        return nil
    }
	definitionList := root.Definitions.AdditionalProperties

	var trees []TreeNode
	// 进行第一遍处理，此时Ref并没有读取，只是记录了引用
    for _, definition := range definitionList {
        str := utils.ToJSON(definition)
        // 解析 Schema 并构建树形结构（中文日志）
        treeRoot, err := parseOpenAPISchema(str, definitionsMap)
        if err != nil {
            klog.V(2).Infof("解析 OpenAPI Schema 失败: %v", err)
            continue
        }
        trees = append(trees, treeRoot)
    }

	docs := &Docs{Trees: trees}
	// 进行遍历处理，将child中ref对应的类型提取出来
	// 此时应该所有的类型都已经存在了
    for i := range docs.Trees {
        docs.loadChild(&docs.Trees[i])
    }
	for i := range docs.Trees {
		docs.loadArrayItems(&docs.Trees[i])
	}
	// 此时 层级结构当中是ref 下面是具体的一个结构体A
	// 结构体A的child是各个属性
	// 我们需要把child下的属性上提一级，避免出现A、再展开才是具体属性的情况

	for i := range docs.Trees {
		docs.childMoveUpLevel(&docs.Trees[i])
	}
	// 将所有节点的ID，改为唯一的

    for i := range docs.Trees {
        docs.uniqueID(&docs.Trees[i])
    }
    return docs
}

// loadArrayItems 递归处理数组类型引用，作为 Docs 的方法
func (d *Docs) loadArrayItems(node *TreeNode) {
	if len(node.Items.Schema) > 0 && node.Items.Schema[0].Ref != "" {
		ref := node.Items.Schema[0].Ref
		if !slice.Contain(blackList, ref) {
			refNode := d.FetchByRef(ref)
			if refNode != nil {
				node.Children = refNode.Children
			}
		}
	}
	for i := range node.Children {
		d.loadArrayItems(node.Children[i])
	}
}

// childMoveUpLevel 属性上提，作为 Docs 的方法
func (d *Docs) childMoveUpLevel(item *TreeNode) {
	name := strings.TrimPrefix(item.Ref, "#/definitions/")
	if item.Ref != "" && len(item.Children) == 1 && item.Children[0].ID == name && len(item.Children[0].Children) > 0 {
		item.Children = item.Children[0].Children
	}
	for i := range item.Children {
		d.childMoveUpLevel(item.Children[i])
	}
}

// loadChild 递归处理引用，作为 Docs 的方法
func (d *Docs) loadChild(item *TreeNode) {
	name := strings.TrimPrefix(item.Ref, "#/definitions/")
	if item.Ref != "" && len(item.Children) > 0 && item.Children[0].ID == name {
		refNode := d.FetchByRef(item.Ref)
		if refNode != nil {
			item.Children[0] = refNode
		}
	}
	for i := range item.Children {
		d.loadChild(item.Children[i])
	}
}

// uniqueID 递归生成唯一ID，作为 Docs 的方法
func (d *Docs) uniqueID(item *TreeNode) {
	item.Value = utils.RandNLengthString(20)
	for i := range item.Children {
		d.uniqueID(item.Children[i])
	}
}

func (d *Docs) ListNames() {
	for _, tree := range d.Trees {
		klog.Infof("tree info ID: %s\tLabel:%s\t\n Parse GVK=[%s,%s,%s]", tree.ID, tree.Label, tree.group, tree.version, tree.kind)
	}
}

// FetchByRef 通过引用查找节点，作为 Docs 的方法
func (d *Docs) FetchByRef(ref string) *TreeNode {
	klog.V(8).Infof("doc FetchByRef: %s", ref)
	id := strings.TrimPrefix(ref, "#/definitions/")
	for _, tree := range d.Trees {
		if tree.ID == id {
			dcp, _ := utils.DeepCopy(tree)
			return &dcp
		}
	}
	return nil
}

func (d *Docs) Fetch(kind string) *TreeNode {
	for _, tree := range d.Trees {
		if tree.Label == kind {
			return &tree
		}
	}
	return nil
}

// FetchByGVK
// com.example.stable.v1.CronTabList
// apiVersion: stable.example.com/v1
// kind: CronTab
func (d *Docs) FetchByGVK(apiVersion, kind string) (node *TreeNode) {
	// 先从 apiVersion+kind 查找，如果找不到再从 kind 查找
	// 采用HasSuffix来匹配,因为内置资源的apiVersion会省略前面的io.k8s.api.core等类似的前缀
	// "id": "io.k8s.api.core.v1.Namespace",
	// group：events.k8s.io =>io.k8s.api.events.v1.Event
	// group""=>io.k8s.api.core.v1.Event
	var group string
	var version string
	if !strings.Contains(apiVersion, "/") {
		group = ""
		version = apiVersion
	} else {
		parts := strings.Split(apiVersion, "/")
		if len(parts) == 2 {
			group = parts[0]
			version = parts[1]
			if strings.Contains(group, ".") {
				ps := strings.Split(group, ".")
				group = ps[0]
			}

		}
	}

	for _, tree := range d.Trees {
		if tree.version == version && tree.kind == kind && tree.group == group {
			node = &tree
			klog.V(6).Infof("[%s:%s]=>[%s,%s,%s]find node ID:%s \tBy GVK(%s,%s,%s)", apiVersion, kind, group, version, kind, tree.ID, tree.group, tree.version, tree.kind)
			return
		}
	}
	for _, tree := range d.Trees {
		if tree.version == version && tree.kind == kind {
			node = &tree
			klog.V(6).Infof("[%s:%s]=>[%s,%s,%s]find node ID:%s \tBy KV(%s,%s)", apiVersion, kind, group, version, kind, tree.ID, tree.version, tree.kind)
			return
		}
	}
	for _, tree := range d.Trees {
		if tree.kind == kind {
			node = &tree
			klog.V(6).Infof("[%s:%s]=>[%s,%s,%s]find node ID:%s \tBy K(%s)", apiVersion, kind, group, version, kind, tree.ID, tree.kind)
			return
		}
	}
	node = d.Fetch(kind)
	klog.V(6).Infof("[%s:%s]=>[%s,%s,%s]find node ID:%s \tBy FetchKind(%s)", apiVersion, kind, group, version, kind, node.ID, node.kind)
	return node
}
